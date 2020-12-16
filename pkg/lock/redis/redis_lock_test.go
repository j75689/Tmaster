package redis

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/j75689/Tmaster/pkg/config"
	"github.com/j75689/Tmaster/pkg/lock"
)

func init() {
	c, err := config.NewConfig("./config/default.config.yaml")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cfg = c
	l, err := NewRedisLocker(
		cfg.Redis.Host,
		cfg.Redis.Port,
		cfg.Redis.Password,
		cfg.Redis.DB,
		cfg.Redis.PoolSize,
		cfg.Redis.MinIdleConns,
		cfg.Redis.LockTimeout,
		cfg.Redis.LockFlushTime,
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	locker = l
}

var (
	cfg    config.Config
	locker lock.Locker
)

func TestRedisLocker_Lock(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name       string
		args       args
		wantIsLock bool
		wantErr    bool
	}{
		{
			name: "Test Redis Lock 1",
			args: args{
				key: "fa123kk13912j",
			},
			wantIsLock: true,
			wantErr:    false,
		},
		{
			name: "Test Redis Lock 2",
			args: args{
				key: "fa123kk13912j",
			},
			wantIsLock: true,
			wantErr:    false,
		},
		{
			name: "Test Redis Lock 3",
			args: args{
				key: "11111",
			},
			wantIsLock: true,
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotIsLock, err := locker.Lock(context.Background(), tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("RedisLocker.Lock() error = %v, wantErr %v", err, tt.wantErr)
				if err := locker.UnLock(context.Background(), tt.args.key); err != nil {
					t.Error(err)
					return
				}
				return
			}
			if gotIsLock != tt.wantIsLock {
				t.Errorf("RedisLocker.Lock() = %v, want %v", gotIsLock, tt.wantIsLock)
				if err := locker.UnLock(context.Background(), tt.args.key); err != nil {
					t.Error(err)
					return
				}
				return
			}
			if err := locker.UnLock(context.Background(), tt.args.key); err != nil {
				t.Error(err)
				return
			}
		})
	}
}

func TestRedisLocker_UnLock(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test Unlock 1",
			args: args{
				key: "123456779",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if islock, err := locker.Lock(context.Background(), tt.args.key); (err != nil) != tt.wantErr || !islock {
				t.Errorf("RedisLocker.UnLock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err := locker.UnLock(context.Background(), tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("RedisLocker.UnLock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestRedisLocker_Delay(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Test Delay 1",
			args: args{
				key: "1233m13i91",
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if islock, err := locker.Lock(context.Background(), tt.args.key); (err != nil) || !islock {
				t.Errorf("RedisLocker.UnLock() error = %v, islock = %v", err, islock)
				return
			}
			got, err := locker.Delay(context.Background(), tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("RedisLocker.Delay() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RedisLocker.Delay() = %v, want %v", got, tt.want)
				return
			}
			if err := locker.UnLock(context.Background(), tt.args.key); err != nil {
				t.Error(err)
				return
			}
		})
	}
}

func TestRedisLocker_LockWithAutoDelay(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Test Redis LockWithAutoDelay 1",
			args: args{
				key: "fa123kk13912j",
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Test Redis LockWithAutoDelay 2",
			args: args{
				key: "fa123kk13912j",
			},
			want:    true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := locker.LockWithAutoDelay(context.Background(), tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("RedisLocker.LockWithAutoDelay() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err := locker.UnLock(context.Background(), tt.args.key); err != nil {
				t.Error(err)
				return
			}
			if got != tt.want {
				t.Errorf("RedisLocker.LockWithAutoDelay() = %v, want %v", got, tt.want)
			}
		})
	}
}
