package redis

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v8"
	"github.com/j75689/Tmaster/pkg/lock"
)

var _ lock.Locker = (*RedisLocker)(nil)

func NewRedisLocker(
	host string, port uint, password string, db int,
	poolSize, minIdleConns int, timeout, flushTime time.Duration,
) (lock.Locker, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", host, port),
		Password:     password,
		DB:           db,
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
	})
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	locker := redislock.New(client)
	redisLocker := &RedisLocker{
		autoDelayLocks: sync.Map{},
		locks:          sync.Map{},
		client:         locker,
		timeout:        timeout,
	}
	redisLocker.flushAutoDelayLock(flushTime)
	return redisLocker, nil
}

type _AutoDelayLock struct {
	*redislock.Lock
	flushTime int64
}

type RedisLocker struct {
	autoDelayLocks sync.Map
	locks          sync.Map
	client         *redislock.Client
	timeout        time.Duration
}

func (locker *RedisLocker) flushAutoDelayLock(flushTime time.Duration) {
	go func() {
		ticker := time.NewTicker(flushTime)
		for range ticker.C {
			locker.autoDelayLocks.Range(func(k, v interface{}) bool {
				if lock, ok := v.(*_AutoDelayLock); ok {
					if time.Now().Unix() > lock.flushTime {
						if err := lock.Refresh(context.Background(), locker.timeout, nil); err == nil {
							lock.flushTime = time.Now().Add(locker.timeout / 2).Unix()
							locker.autoDelayLocks.Store(k, lock)
						}
					}
				}
				return true
			})
		}
	}()
}

func (locker *RedisLocker) checkCtx(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	ctx, _ = context.WithTimeout(context.Background(), locker.timeout)
	return ctx
}

func (locker *RedisLocker) LockWithAutoDelay(ctx context.Context, key string) (bool, error) {
	ctx = locker.checkCtx(ctx)
	lock, err := locker.client.Obtain(ctx, key, locker.timeout, nil)
	if err != nil {
		return false, err
	}
	locker.autoDelayLocks.Store(key, &_AutoDelayLock{
		Lock:      lock,
		flushTime: time.Now().Add(locker.timeout / 2).Unix(),
	})
	return true, nil
}

func (locker *RedisLocker) Lock(ctx context.Context, key string) (bool, error) {
	ctx = locker.checkCtx(ctx)
	lock, err := locker.client.Obtain(ctx, key, locker.timeout, nil)
	if err != nil {
		return false, err
	}
	locker.locks.Store(key, lock)
	return true, nil
}

func (locker *RedisLocker) UnLock(ctx context.Context, key string) error {
	ctx = locker.checkCtx(ctx)
	if v, ok := locker.locks.Load(key); ok {
		if lock, ok := v.(*redislock.Lock); ok {
			defer locker.locks.Delete(key)
			return lock.Release(ctx)
		}
	}
	if v, ok := locker.autoDelayLocks.Load(key); ok {
		if lock, ok := v.(*_AutoDelayLock); ok {
			defer locker.autoDelayLocks.Delete(key)
			return lock.Release(ctx)
		}
	}
	return fmt.Errorf("lock [%s] not found", key)
}

func (locker *RedisLocker) Delay(ctx context.Context, key string) (bool, error) {
	ctx = locker.checkCtx(ctx)
	if v, ok := locker.locks.Load(key); ok {
		if lock, ok := v.(*redislock.Lock); ok {
			if err := lock.Refresh(ctx, locker.timeout, nil); err != nil {
				return false, err
			}
			return true, nil
		}
	}
	return false, fmt.Errorf("lock [%s] not found", key)
}
