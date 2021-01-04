package stream

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/adjust/rmq/v4"
	redis "github.com/go-redis/redis/v8"
	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/mq"
	"github.com/rs/zerolog"
)

var _ mq.MQ = (*RedisStream)(nil)

// NewRedisStream returns a mq.MQ implement with redis
func NewRedisStream(config config.MQConfig, logger zerolog.Logger) (mq.MQ, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.RedisStream.RedisOption.Host, config.RedisStream.RedisOption.Port),
		Password:     config.RedisStream.RedisOption.Password,
		DB:           config.RedisStream.RedisOption.DB,
		PoolSize:     config.RedisStream.RedisOption.PoolSize,
		MinIdleConns: config.RedisStream.RedisOption.MinIdleConns,
	})
	ctx, cancel := context.WithTimeout(context.Background(), config.RedisStream.RedisOption.DialTimeout)
	defer cancel()
	_, err = client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	conn, err := rmq.OpenConnectionWithRedisClient(hostname+strconv.Itoa(rand.Int()), client, nil)
	if err != nil {
		return nil, err
	}
	return &RedisStream{
		config:     config,
		logger:     logger,
		conn:       conn,
		publishers: &sync.Map{},
		consumers:  &sync.Map{},
	}, nil
}

type RedisStream struct {
	config     config.MQConfig
	logger     zerolog.Logger
	conn       rmq.Connection
	publishers *sync.Map
	consumers  *sync.Map
}

func (stream *RedisStream) getQueue(pool *sync.Map, key string) (rmq.Queue, error) {
	if v, ok := pool.Load(key); ok {
		if client, ok := v.(rmq.Queue); ok {
			return client, nil
		}
	}

	client, err := stream.conn.OpenQueue(key)
	if err != nil {
		return nil, fmt.Errorf("redis stream new queue error: %v", err)
	}
	defer pool.Store(key, client)

	return client, nil
}

func (stream *RedisStream) getPublisher(projectID, topicID string) (rmq.Queue, error) {
	return stream.getQueue(stream.publishers, projectID+"."+topicID)
}

func (stream *RedisStream) getConsumers(projectID, subscription string) (rmq.Queue, error) {
	return stream.getQueue(stream.publishers, projectID+"."+subscription)
}

func (stream *RedisStream) InitTopic(ctx context.Context, projectID, topicID string) error {
	// RedisStream does not need to be initialized
	return nil
}
func (stream *RedisStream) InitSubscriber(ctx context.Context, projectID, topicID string, subIDs ...string) error {
	// RedisStream does not need to be initialized
	return nil
}
func (stream *RedisStream) Subscribe(projectID, subscription string, process func(context.Context, []byte) error) error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	queue, err := stream.getConsumers(projectID, subscription)
	if err != nil {
		return err
	}
	if err := queue.StartConsuming(stream.config.RedisStream.PrefetchLimit, stream.config.RedisStream.PollDuration); err != nil {
		return err
	}
	stream.logger.Info().Msg("start subscribe: " + projectID + "." + subscription)
	for i := 0; i < stream.config.RedisStream.NumConsumers; i++ {
		if _, err := queue.AddConsumer(
			hostname+strconv.Itoa(rand.Int())+"-"+strconv.Itoa(i),
			NewConsumer(projectID+"."+subscription, stream.config.RedisStream.ProcessTimeout, stream.logger, process),
		); err != nil {
			return err
		}
	}
	select {}
}
func (stream *RedisStream) Publish(distributedID int64, projectID, topicID string, message []byte) error {
	t := time.Now()
	defer stream.logger.Debug().
		Str("queue", projectID+"."+topicID).
		Int("size", len(message)).
		Dur("duration", time.Since(t)).
		Msg("publish msg")
	queue, err := stream.getPublisher(projectID, topicID)
	if err != nil {
		return err
	}

	return queue.PublishBytes(message)
}
func (stream *RedisStream) Stop() {
	<-stream.conn.StopAllConsuming()
}

type Consumer struct {
	queue          string
	processTimeout time.Duration
	logger         zerolog.Logger
	process        func(context.Context, []byte) error
}

func NewConsumer(queue string, processTimeout time.Duration, logger zerolog.Logger, process func(context.Context, []byte) error) *Consumer {
	return &Consumer{
		queue:          queue,
		processTimeout: processTimeout,
		logger:         logger,
		process:        process,
	}
}

func (consumer *Consumer) Consume(delivery rmq.Delivery) {
	ctx, cancel := context.WithTimeout(context.Background(), consumer.processTimeout)
	defer cancel()
	if err := consumer.process(ctx, []byte(delivery.Payload())); err != nil {
		consumer.logger.
			Err(err).
			Str("queue", consumer.queue).
			Msg("process message error")
	}
	if err := delivery.Ack(); err != nil {
		consumer.logger.
			Err(err).
			Str("queue", consumer.queue).
			Msg("error on subscribe ack")
	}
}
