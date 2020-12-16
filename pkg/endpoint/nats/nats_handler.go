package nats

import (
	"context"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"github.com/rs/zerolog"
	"github.com/j75689/Tmaster/pkg/endpoint"
	"github.com/j75689/Tmaster/pkg/graph/model"
)

type _NatsClientKey struct {
	Servers     string
	Subject     string
	User        string
	Password    string
	Token       string
	ClusterName string
}

var _ endpoint.Handler = (*NatsHandler)(nil)

func NewNatsHandler(logger zerolog.Logger) *NatsHandler {
	handler := &NatsHandler{
		logger:     logger,
		clientPool: &sync.Map{},
	}
	handler.flushConn()
	return handler
}

type NatsHandler struct {
	logger     zerolog.Logger
	clientPool *sync.Map
}

func (handler *NatsHandler) flushConn() {
	go func() {
		ticker := time.NewTicker(time.Minute)
		for range ticker.C {
			handler.clientPool.Range(func(k, v interface{}) bool {
				if client, ok := v.(_NatsClient); ok {
					if time.Now().UTC().Unix() > client.FlushTime() {
						client.Close()
						handler.clientPool.Delete(k)
					}
				}
				return true
			})
		}
	}()
}

func (handler *NatsHandler) getClient(servers, subject, user, password, token, clusterName string) (_NatsClient, error) {
	key := _NatsClientKey{
		Servers:     servers,
		Subject:     subject,
		User:        user,
		Password:    password,
		Token:       token,
		ClusterName: clusterName,
	}
	if v, ok := handler.clientPool.Load(key); ok {
		if client, ok := v.(_NatsClient); ok {
			client.SetFlushTime(time.Now().UTC().Add(time.Hour).Unix())
			handler.clientPool.Store(key, client)
			return client, nil
		}
	}
	var client _NatsClient
	serverUrls := servers
	opts := []nats.Option{}
	if user != "" && password != "" {
		opts = append(opts, nats.UserInfo(user, password))
	}
	if token != "" {
		opts = append(opts, nats.Token(token))
	}
	conn, err := nats.Connect(serverUrls, opts...)
	if err != nil {
		return nil, err
	}
	client = &_NatsNormalClient{conn: conn}

	// streaming mode
	if clusterName != "" {
		sconn, err := stan.Connect(clusterName, strconv.Itoa(rand.Int()), stan.NatsConn(conn),
			stan.Pings(1, 2),
			stan.SetConnectionLostHandler(func(_ stan.Conn, err error) {
				handler.clientPool.Delete(key)
			}))
		if err != nil {
			return nil, err
		}
		client = &_NatsStanClient{conn: sconn}
	}

	defer func() {
		client.SetFlushTime(time.Now().UTC().Add(time.Hour).Unix())
		handler.clientPool.Store(key, client)
	}()
	return client, nil
}

func (handler *NatsHandler) Do(ctx context.Context, endpointConfig *model.Endpoint) (map[string]string, interface{}, error) {
	handler.logger.Debug().Msg("start nats endpoint")
	defer handler.logger.Debug().Msg("complete nats endpoint")
	handler.logger.Debug().Interface("body", endpointConfig).Msgf("endpoint config")

	var (
		servers, subject, user, password, token, clusterName string
		body                                                 string
	)
	if endpointConfig.URL != nil {
		servers = *endpointConfig.URL
	}
	if endpointConfig.Subject != nil {
		subject = *endpointConfig.Subject
	}
	if endpointConfig.User != nil {
		user = *endpointConfig.User
	}
	if endpointConfig.Password != nil {
		password = *endpointConfig.Password
	}
	if endpointConfig.Token != nil {
		token = *endpointConfig.Token
	}
	if endpointConfig.ClusterName != nil {
		clusterName = *endpointConfig.ClusterName
	}
	if endpointConfig.Body != nil {
		body = *endpointConfig.Body
	}

	client, err := handler.getClient(servers, subject, user, password, token, clusterName)
	if err != nil {
		return nil, nil, err
	}

	return nil, nil, client.Publish(subject, []byte(body))
}
