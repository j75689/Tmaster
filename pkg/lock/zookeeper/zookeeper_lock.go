package zookeeper

import (
	"context"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/j75689/Tmaster/pkg/lock"
)

var _ lock.Locker = (*ZookeeperLock)(nil)

func NewZookeeperLock() lock.Locker {
	zkConn, events, err := zk.Connect([]string{}, time.Second)
	_ = events
	_ = zkConn
	_ = err
	return &ZookeeperLock{}
}

type ZookeeperLock struct {
}

func (lock *ZookeeperLock) LockWithAutoDelay(ctx context.Context, key string) (bool, error) {
	return false, nil
}
func (lock *ZookeeperLock) Lock(ctx context.Context, key string) (bool, error) {
	return false, nil
}
func (lock *ZookeeperLock) UnLock(ctx context.Context, key string) error {
	return nil
}
func (lock *ZookeeperLock) Delay(ctx context.Context, key string) (bool, error) {
	return false, nil
}
