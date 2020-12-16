package lock

import "context"

type Locker interface {
	LockWithAutoDelay(ctx context.Context, key string) (bool, error)
	Lock(ctx context.Context, key string) (bool, error)
	UnLock(ctx context.Context, key string) error
	Delay(ctx context.Context, key string) (bool, error)
}
