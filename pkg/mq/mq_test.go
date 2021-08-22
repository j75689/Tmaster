package mq

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/j75689/Tmaster/mock"
)

func TestPublish(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockMQ := mock.NewMockMQ(controller)
	mockMQ.
		EXPECT().
		Publish(int64(0), "test", "test", []byte("test")).
		Return(nil).
		AnyTimes()

	if err := mockMQ.Publish(int64(0), "test", "test", []byte("test")); err != nil {
		t.Error(err)
	}
}

func TestSubscribe(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	callback := func(context.Context, []byte) error {
		return nil
	}

	mockMQ := mock.NewMockMQ(controller)
	mockMQ.
		EXPECT().
		Subscribe("test", "test", gomock.AssignableToTypeOf(callback)).
		Return(nil).
		AnyTimes()

	if err := mockMQ.Subscribe("test", "test", callback); err != nil {
		t.Error(err)
	}
}
