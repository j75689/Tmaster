package redisstream

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"

	"github.com/adjust/rmq/v4"
	"github.com/go-redis/redis/v8"
	"github.com/j75689/Tmaster/pkg/endpoint"
	"github.com/j75689/Tmaster/pkg/graph/model"
	"github.com/rs/zerolog"
)

type _RedisConnectionKey struct {
	Addr     string
	Password string
	DB       int
}

type _QueueKey struct {
	_RedisConnectionKey
	Queue string
}

var _ endpoint.Handler = (*RedisStreamHandler)(nil)

func NewRedisStreamHandler(logger zerolog.Logger) *RedisStreamHandler {
	return &RedisStreamHandler{
		logger:    logger,
		connPool:  &sync.Map{},
		queuePool: &sync.Map{},
	}
}

type RedisStreamHandler struct {
	logger    zerolog.Logger
	connPool  *sync.Map
	queuePool *sync.Map
}

func (handler *RedisStreamHandler) getQueue(ctx context.Context, addr, password string, db int, queue string) (rmq.Queue, error) {
	key := _QueueKey{
		_RedisConnectionKey: _RedisConnectionKey{
			Addr:     addr,
			Password: password,
			DB:       db,
		},
		Queue: queue,
	}
	if v, ok := handler.queuePool.Load(key); ok {
		if queue, ok := v.(rmq.Queue); ok {
			return queue, nil
		}
	}

	conn, err := handler.getConn(ctx, addr, password, db)
	if err != nil {
		return nil, err
	}

	q, err := conn.OpenQueue(queue)
	if err != nil {
		return nil, fmt.Errorf("redis stream handler new queue error: %v", err)
	}
	defer handler.queuePool.Store(key, q)
	return q, nil
}

func (handler *RedisStreamHandler) getConn(ctx context.Context, addr, password string, db int) (rmq.Connection, error) {
	key := _RedisConnectionKey{
		Addr:     addr,
		Password: password,
		DB:       db,
	}
	if v, ok := handler.connPool.Load(key); ok {
		if client, ok := v.(rmq.Connection); ok {
			return client, nil
		}
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	_, err = client.WithContext(ctx).Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("redis stream handler ping redis error: %v", err)
	}

	conn, err := rmq.OpenConnectionWithRedisClient(hostname+strconv.Itoa(rand.Int()), client, nil)
	if err != nil {
		return nil, fmt.Errorf("redis stream handler new connection error: %v", err)
	}

	defer handler.connPool.Store(key, conn)
	return conn, nil
}

func (handler *RedisStreamHandler) Do(ctx context.Context, endpointConfig *model.Endpoint) (map[string]string, interface{}, error) {
	handler.logger.Debug().Msg("start redis stream endpoint")
	defer handler.logger.Debug().Msg("complete redis stream endpoint")
	handler.logger.Debug().Interface("body", endpointConfig).Msg("endpoint config")

	var (
		addr, password, queue, body string
		db                          int
	)
	if endpointConfig.URL != nil {
		addr = *endpointConfig.URL
	}
	if endpointConfig.Password != nil {
		password = *endpointConfig.Password
	}
	if endpointConfig.Queue != nil {
		queue = *endpointConfig.Queue
	}
	if endpointConfig.Db != nil {
		db = *endpointConfig.Db
	}
	if endpointConfig.Body != nil {
		body = *endpointConfig.Body
	}

	q, err := handler.getQueue(ctx, addr, password, db, queue)
	if err != nil {
		return nil, nil, err
	}
	return nil, true, q.Publish(body)
}
