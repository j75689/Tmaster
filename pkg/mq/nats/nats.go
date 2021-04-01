package nats

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/mq"
	nats "github.com/nats-io/nats.go"
	stan "github.com/nats-io/stan.go"
	"github.com/rs/zerolog"
)

var _ mq.MQ = (*Nats)(nil)

// NewNats returns a mq.MQ implement with nats streaming
func NewNats(config config.MQConfig, logger zerolog.Logger) (mq.MQ, error) {
	mq := &Nats{
		logger:      logger,
		config:      config,
		ncConn:      make([]*nats.Conn, config.Distribution),
		stanConn:    make([]stan.Conn, config.Distribution),
		subscribers: &sync.Map{},
	}

	for i := 0; i < config.Distribution; i++ {
		if err := mq.initNatsConnection(i); err != nil {
			return nil, err
		}
	}

	return mq, nil
}

type Nats struct {
	logger      zerolog.Logger
	ncConn      []*nats.Conn
	stanConn    []stan.Conn
	subscribers *sync.Map
	config      config.MQConfig
}

func (nat *Nats) initNatsConnection(distribution int) error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	serverUrls := strings.Join(nat.config.Nats.Servers, ", ")
	opts := []nats.Option{
		nats.Name(hostname + "-" + strconv.Itoa(distribution)),
		nats.MaxReconnects(nat.config.Nats.MaxReconnect),
		nats.ReconnectWait(nat.config.Nats.ReconnectWait),
		nats.RetryOnFailedConnect(nat.config.Nats.RetryOnFailedConnect),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			nat.logger.Info().Msg("nats client reconnected")
		}),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			nat.logger.Err(err).Msg("nats client disconnect")
		}),
	}
	if nat.config.Nats.User != "" && nat.config.Nats.Password != "" {
		opts = append(opts, nats.UserInfo(nat.config.Nats.User, nat.config.Nats.Password))
	}
	if nat.config.Nats.Token != "" {
		opts = append(opts, nats.Token(nat.config.Nats.Token))
	}
	conn, err := nats.Connect(serverUrls, opts...)
	if err != nil {
		return err
	}

	nat.ncConn[distribution] = conn
	return nat.initStanConnection(distribution)
}

func (nat *Nats) initStanConnection(distribution int) error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	for range time.NewTicker(nat.config.Nats.ReconnectWait).C {
		sc, err := stan.Connect(nat.config.Nats.ClusterName, hostname+strconv.Itoa(rand.Int())+"-"+strconv.Itoa(distribution),
			stan.NatsConn(nat.ncConn[distribution]),
			stan.Pings(nat.config.Nats.PingInterval, nat.config.Nats.PingMaxOut),
			stan.MaxPubAcksInflight(nat.config.Nats.MaxPubAcksInflight),
			stan.SetConnectionLostHandler(func(_ stan.Conn, err error) {
				nat.logger.Err(err).Msgf("stan connection lost")
				nat.subscribers.Range(func(k, v interface{}) bool {
					if sub, ok := v.(*Subscriber); ok {
						sub.Stop()
					}
					return true
				})
				nat.initStanConnection(distribution)
			}),
		)
		if err != nil {
			nat.logger.Err(err).Msgf("connect stan server error")
			continue
		}
		nat.stanConn[distribution] = sc
		nat.subscribers.Range(func(k, v interface{}) bool {
			if sub, ok := v.(*Subscriber); ok {
				if err := sub.Start(sc); err != nil {
					nat.logger.Err(err).Msg("restart subscriber error")
					nat.subscribers.Delete(k)
				}
			}
			return true
		})
		return nil
	}
	return nil
}

func (nat *Nats) InitTopic(ctx context.Context, projectID, topicID string) error {
	// Nats does not need to be initialized
	return nil
}

func (nat *Nats) InitSubscriber(ctx context.Context, projectID, topicID string, subIDs ...string) error {
	// Nats does not need to be initialized
	return nil
}

func (nat *Nats) Subscribe(projectID, subscription string, process func(context.Context, []byte) error) error {
	for i, group := range _GetGroups(nat.config.Distribution) {
		sub := NewSubscriber(
			nat.logger,
			nat.config.Nats.QueueGroup+"-"+group.String(),
			projectID,
			subscription+"."+group.String(),
			process,
			nat.config.Nats.DurableName,
			nat.config.Nats.AckWait,
			nat.config.Nats.MaxInflight,
			nat.config.Nats.WorkerSize,
		)
		if err := sub.Start(nat.stanConn[i]); err != nil {
			return err
		}
		nat.subscribers.Store(projectID+"."+subscription+"-"+group.String(), sub)
	}

	select {}
}

func (nat *Nats) Publish(distributedID int64, projectID, topicID string, message []byte) error {
	t := time.Now()
	idx, group := _GetDistributedGroupName(nat.config.Distribution, distributedID)
	_, err := nat.stanConn[idx].PublishAsync(projectID+"."+topicID+"."+group.String(),
		message, func(lguid string, err error) {
			if err != nil {
				nat.logger.Err(err).Str("lguid", lguid).Msg("Publisher got following error")
				return
			}
			nat.logger.Debug().
				Str("subject", projectID+"."+topicID+"."+group.String()).
				Int("size", len(message)).
				Str("lguid", lguid).
				Dur("duration", time.Since(t)).
				Msg("publish msg")
		})

	return err
}

func (nat *Nats) Stop() {
	nat.subscribers.Range(func(k, v interface{}) bool {
		if sub, ok := v.(*Subscriber); ok {
			sub.Stop()
		}
		nat.subscribers.Delete(k)
		return true
	})
	for i := 0; i < nat.config.Distribution; i++ {
		nat.stanConn[i].Close()
		if !nat.ncConn[i].IsClosed() {
			nat.ncConn[i].Close()
		}
	}
}

func NewSubscriber(
	logger zerolog.Logger,
	queueGroup, projectID, subscription string,
	process func(context.Context, []byte) error,
	durableName string,
	ackWait time.Duration,
	maxInflight, workerSize int,
) *Subscriber {
	return &Subscriber{
		logger:       logger,
		queueGroup:   queueGroup,
		projectID:    projectID,
		subscription: subscription,
		process:      process,
		workerSize:   workerSize,
		durableName:  durableName,
		ackWait:      ackWait,
		maxInflight:  maxInflight,
	}
}

type Subscriber struct {
	logger        zerolog.Logger
	queueGroup    string
	projectID     string
	subscription  string
	process       func(context.Context, []byte) error
	durableName   string
	ackWait       time.Duration
	maxInflight   int
	workerSize    int
	workerChannel chan *stan.Msg
	cancel        context.CancelFunc
}

func (sub *Subscriber) Start(stanConn stan.Conn) error {
	sub.logger.Info().Msg("start subscribe: " + sub.projectID + "." + sub.subscription)
	sub.workerChannel = make(chan *stan.Msg)
	s, err := stanConn.QueueSubscribe(sub.projectID+"."+sub.subscription, sub.queueGroup, func(m *stan.Msg) {
		sub.workerChannel <- m
	},
		stan.StartWithLastReceived(),
		stan.SetManualAckMode(),
		stan.DurableName(sub.durableName),
		stan.AckWait(sub.ackWait),
		stan.MaxInflight(sub.maxInflight),
	)
	if err != nil {
		return err
	}
	err = s.SetPendingLimits(sub.workerSize, -1)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	for i := 0; i < sub.workerSize; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case m := <-sub.workerChannel:
					sub.logger.Info().Str("subject", m.Subject).Msg("recived")
					if err := sub.process(context.Background(), m.Data); err != nil {
						sub.logger.Err(err).Str("subject", m.Subject).Bytes("data", m.Data).Msg(
							"error on subscribe",
						)
					}
					if err := m.Ack(); err != nil {
						sub.logger.Err(err).Str("subject", m.Subject).Msg(
							"error on subscribe ack",
						)
					}
				}
			}
		}()
	}
	sub.cancel = cancel
	return nil
}

func (sub *Subscriber) Stop() {
	if sub.cancel != nil {
		sub.cancel()
	}
	sub.logger.Info().Msg("stop subscribe: " + sub.projectID + "." + sub.subscription)
}
