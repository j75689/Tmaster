package stream

import (
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/mq"
)

func initLocalRedisMQ() (mq.MQ, error) {
	return NewRedisStream(
		config.MQConfig{
			Driver: "redis_stream",
			RedisStream: config.RedisStreamArg{
				RedisOption: config.RedisConfig{
					Host:         "localhost",
					Port:         6379,
					DB:           1,
					PoolSize:     1000,
					MinIdleConns: 100,
				},
			},
		},
		log.Logger,
	)
}

func BenchmarkPublish(b *testing.B) {
	mq, err := initLocalRedisMQ()
	if err != nil {
		b.Error(err)
		return
	}
	for i := 0; i < b.N; i++ {
		mq.Publish(time.Now().Unix(), "test", "BenchmarkPublish", []byte("hello"))
	}
}
