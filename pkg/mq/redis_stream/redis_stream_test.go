package stream

import (
	"context"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/adjust/rmq/v4"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/mq"
	"github.com/rs/zerolog/log"
)

func newMockRedisStream() (mq.MQ, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	mr, err := miniredis.Run()
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
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
		config: config.MQConfig{
			RedisStream: config.RedisStreamArg{
				NumConsumers:   100,
				PrefetchLimit:  100,
				PollDuration:   time.Millisecond,
				ProcessTimeout: 1 * time.Second,
			},
		},
		logger:     log.Logger,
		conn:       conn,
		publishers: &sync.Map{},
		consumers:  &sync.Map{},
	}, nil
}

func BenchmarkPublish(b *testing.B) {
	mq, err := newMockRedisStream()
	if err != nil {
		b.Error(err)
		return
	}
	for i := 0; i < b.N; i++ {
		mq.Publish(time.Now().Unix(), "test", "BenchmarkPublish", []byte("hello"))
	}
}
