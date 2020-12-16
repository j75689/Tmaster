package wireset

import (
	"github.com/google/wire"
	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/lock"
	"github.com/j75689/Tmaster/pkg/lock/redis"
)

var LockerSet = wire.NewSet(
	InitializeLocker,
)

func InitializeLocker(config config.Config) (lock.Locker, error) {
	return redis.NewRedisLocker(
		config.Redis.Host, config.Redis.Port, config.Redis.Password,
		config.Redis.DB, config.Redis.PoolSize,
		config.Redis.MinIdleConns, config.Redis.LockTimeout, config.Redis.LockFlushTime,
	)
}
