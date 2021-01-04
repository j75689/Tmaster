package pubsub

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/mq"
	"github.com/rs/zerolog"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ mq.MQ = (*GooglePubSub)(nil)

// NewGooglePubSub returns a mq.MQ implement with google pubsub
func NewGooglePubSub(config config.MQConfig, logger zerolog.Logger) (mq.MQ, error) {
	credentialPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if config.GooglePubSub.CredentialPath != "" {
		credentialPath = config.GooglePubSub.CredentialPath
	}
	return &GooglePubSub{
		topicPool:              &sync.Map{},
		publishedClientPool:    &sync.Map{},
		subscribeClientPool:    &sync.Map{},
		logger:                 logger,
		credentialPath:         credentialPath,
		stopTimeout:            config.StopTimeout,
		synchronous:            config.GooglePubSub.Synchronous,
		maxOutstandingMessages: config.GooglePubSub.MaxOutstandingMessages,
		maxOutstandingBytes:    config.GooglePubSub.MaxOutstandingBytes,
		numGoroutines:          config.GooglePubSub.NumGoroutines,
	}, nil
}

// GooglePubSub implements methods of MQ through google pub/sub
type GooglePubSub struct {
	topicPool              *sync.Map
	publishedClientPool    *sync.Map
	subscribeClientPool    *sync.Map
	logger                 zerolog.Logger
	credentialPath         string
	receiver               []*Receiver
	stopTimeout            time.Duration
	synchronous            bool
	maxOutstandingMessages int
	maxOutstandingBytes    int
	numGoroutines          int
}

func (googlepubsub *GooglePubSub) getTopic(client *pubsub.Client, projectID, topicID string) *pubsub.Topic {
	if v, ok := googlepubsub.topicPool.Load(topicID); ok {
		if topic, ok := v.(*pubsub.Topic); ok {
			return topic
		}
	}

	topic := client.TopicInProject(topicID, projectID)
	defer googlepubsub.topicPool.Store(topicID, topic)
	return topic
}

func (googlepubsub *GooglePubSub) getPublishedClient(ctx context.Context, projectID string) (*pubsub.Client, error) {
	return googlepubsub.getClient(ctx, projectID, googlepubsub.publishedClientPool)
}

func (googlepubsub *GooglePubSub) getSubscribeClient(ctx context.Context, projectID string) (*pubsub.Client, error) {
	return googlepubsub.getClient(ctx, projectID, googlepubsub.subscribeClientPool)
}

func (googlepubsub *GooglePubSub) getClient(ctx context.Context, projectID string, pool *sync.Map) (*pubsub.Client, error) {
	if v, ok := pool.Load(projectID); ok {
		if client, ok := v.(*pubsub.Client); ok {
			return client, nil
		}
	}

	client, err := pubsub.NewClient(ctx, projectID,
		option.WithCredentialsFile(googlepubsub.credentialPath),
	)
	if err != nil {
		return nil, fmt.Errorf("pubsub new client error: %v", err)
	}
	defer pool.Store(projectID, client)

	return client, nil
}

func (googlepubsub *GooglePubSub) InitTopic(ctx context.Context, projectID, topicID string) error {
	client, err := googlepubsub.getPublishedClient(ctx, projectID)
	if err != nil {
		return err
	}

	topic := client.TopicInProject(topicID, projectID)
	exist, err := topic.Exists(ctx)
	if err != nil {
		return err
	}

	if exist {
		return nil
	}

	_, err = client.CreateTopic(ctx, topicID)
	return err
}
func (googlepubsub *GooglePubSub) InitSubscriber(ctx context.Context, projectID, topicID string, subIDs ...string) error {
	client, err := googlepubsub.getPublishedClient(ctx, projectID)
	if err != nil {
		return err
	}

	topic := client.TopicInProject(topicID, projectID)
	exist, err := topic.Exists(ctx)
	if err != nil {
		return err
	}

	if !exist {
		return fmt.Errorf("topic: %s not exist", topicID)
	}

	// Create a new subscription to the previously
	// created topic and ensure it never expires.
	for _, subID := range subIDs {
		sub := client.SubscriptionInProject(subID, projectID)
		exist, err := sub.Exists(ctx)
		if err != nil {
			return err
		}

		if exist {
			continue
		}

		_, err = client.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{
			Topic:            topic,
			AckDeadline:      30 * time.Second,
			ExpirationPolicy: time.Duration(0),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (googlepubsub *GooglePubSub) Subscribe(projectID, subscription string, process func(context.Context, []byte) error) error {
	ctx := context.Background()
	client, err := googlepubsub.getSubscribeClient(ctx, projectID)
	if err != nil {
		return err
	}

	sub := client.SubscriptionInProject(subscription, projectID)
	sub.ReceiveSettings.Synchronous = googlepubsub.synchronous
	// MaxOutstandingMessages is the maximum number of unprocessed messages the
	// client will pull from the server before pausing.
	//
	// This is only guaranteed when ReceiveSettings.Synchronous is set to true.
	// When Synchronous is set to false, the StreamingPull RPC is used which
	// can pull a single large batch of messages at once that is greater than
	// MaxOustandingMessages before pausing. For more info, see
	// https://cloud.google.com/pubsub/docs/pull#streamingpull_dealing_with_large_backlogs_of_small_messages.
	sub.ReceiveSettings.MaxOutstandingMessages = googlepubsub.maxOutstandingMessages

	// MaxOutstandingBytes is the maximum size of unprocessed messages,
	// that the client will pull from the server before pausing. Similar
	// to MaxOutstandingMessages, this may be exceeded with a large batch
	// of messages since we cannot control the size of a batch of messages
	// from the server (even with the synchronous Pull RPC).
	sub.ReceiveSettings.MaxOutstandingBytes = googlepubsub.maxOutstandingBytes

	// NumGoroutines is the number of goroutines Receive will spawn to pull
	// messages concurrently. If NumGoroutines is less than 1, it will be treated
	// as if it were DefaultReceiveSettings.NumGoroutines.
	//
	// NumGoroutines does not limit the number of messages that can be processed
	// concurrently. Even with one goroutine, many messages might be processed at
	// once, because that goroutine may continually receive messages and invoke the
	// function passed to Receive on them. To limit the number of messages being
	// processed concurrently, set MaxOutstandingMessages.
	sub.ReceiveSettings.NumGoroutines = googlepubsub.numGoroutines
	googlepubsub.receiver = append(googlepubsub.receiver, NewReceiver(sub, process, googlepubsub.logger, googlepubsub.stopTimeout))
	return nil
}

func (googlepubsub *GooglePubSub) Publish(distributedID int64, projectID, topicID string, message []byte) error {
	ctx := context.Background()
	client, err := googlepubsub.getPublishedClient(ctx, projectID)
	if err != nil {
		return err
	}

	t := googlepubsub.getTopic(client, projectID, topicID)
	result := t.Publish(ctx, &pubsub.Message{
		Data: []byte(message),
	})

	_, err = result.Get(ctx)
	return err
}

func (googlepubsub *GooglePubSub) Stop() {
	for _, receiver := range googlepubsub.receiver {
		receiver.Stop()
	}
	googlepubsub.publishedClientPool.Range(func(k, v interface{}) bool {
		if client, ok := v.(*pubsub.Client); ok {
			client.Close()
		}
		return true
	})
	googlepubsub.subscribeClientPool.Range(func(k, v interface{}) bool {
		if client, ok := v.(*pubsub.Client); ok {
			client.Close()
		}
		return true
	})
}

func NewReceiver(subscription *pubsub.Subscription,
	process func(context.Context, []byte) error,
	logger zerolog.Logger,
	stopTimeout time.Duration) *Receiver {
	ctx, cancel := context.WithCancel(context.Background())
	receiver := &Receiver{
		logger:       logger,
		subscription: subscription,
		process:      process,
		cancel:       cancel,
		notifyDone:   make(chan struct{}),
		stopTimeout:  stopTimeout,
		startFailed:  make(chan struct{}),
	}

	return receiver.Start(ctx)
}

type Receiver struct {
	sync.Mutex
	logger       zerolog.Logger
	stopTimeout  time.Duration
	subscription *pubsub.Subscription
	process      func(context.Context, []byte) error
	cancel       context.CancelFunc
	notifyDone   chan struct{}
	unAckMsg     uint
	stopReceive  bool
	startFailed  chan struct{}
}

func (receiver *Receiver) Start(ctx context.Context) *Receiver {
	go receiver.monitoring(ctx)
	return receiver.receive(ctx)
}

func (receiver *Receiver) Stop() {
	receiver.cancel()
	<-receiver.notifyDone
}

func (receiver *Receiver) receive(ctx context.Context) *Receiver {
	receiver.logger.Info().Msgf("start subscribe: %s", receiver.subscription.ID())
	err := receiver.subscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		if receiver.stopReceive {
			msg.Nack()
			return
		}
		receiver.Lock()
		receiver.unAckMsg++
		receiver.Unlock()
		if err := receiver.process(ctx, msg.Data); err != nil {
			receiver.logger.Err(err).Str("message_id", msg.ID).Bytes("data", msg.Data).Msgf(
				"error on subscribe: %s", receiver.subscription.ID(),
			)
		}
		msg.Ack()
		receiver.Lock()
		receiver.unAckMsg--
		receiver.Unlock()
	})
	if err != nil {
		if statusErr, _ := status.FromError(err); statusErr.Code() != codes.Canceled {
			receiver.logger.Err(err).Msg("start subscribe failed")
		}
		close(receiver.startFailed)
	}

	return receiver
}

func (receiver *Receiver) monitoring(ctx context.Context) {
	select {
	case <-ctx.Done():
		receiver.stopReceive = true
		receiver.logger.Info().Msg(fmt.Sprintf("stop subscription: %s", receiver.subscription.ID()))
	case <-receiver.startFailed:
		receiver.stopReceive = true
		close(receiver.notifyDone)
		return
	}

	complete := make(chan struct{})
	go func() {
		// wait for all messages to be acked
		for receiver.isRunning() {
			time.Sleep(time.Second)
		}
		close(complete)
	}()

	select {
	case <-complete:
		receiver.logger.Info().Msg(fmt.Sprintf("subscribe: %s, all message acked", receiver.subscription.ID()))
	case <-time.After(receiver.stopTimeout):
		// if timeout, forcibly stop the process
		receiver.logger.Info().Msg(fmt.Sprintf("subscribe: %s, forcibly stop", receiver.subscription.ID()))
	}

	close(receiver.notifyDone)
}

func (receiver *Receiver) isRunning() bool {
	return receiver.unAckMsg > 0
}
