package mq

import (
	"context"
)

type MQ interface {
	InitTopic(ctx context.Context, projectID, topicID string) error
	InitSubscriber(ctx context.Context, projectID, topicID string, subIDs ...string) error
	Subscribe(projectID, subscription string, process func(context.Context, []byte) error) error
	Publish(distributedID int64, projectID, topicID string, message []byte) error
	Stop()
}
