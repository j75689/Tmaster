package pubsub

import (
	"context"
	"fmt"
	"sync"

	"cloud.google.com/go/pubsub"
	"github.com/rs/zerolog"
	"github.com/j75689/Tmaster/pkg/endpoint"
	"github.com/j75689/Tmaster/pkg/graph/model"
	"google.golang.org/api/option"
)

type _PubSubClientKey struct {
	ProjectID  string
	Credential string
}

var _ endpoint.Handler = (*PubSubHandler)(nil)

func NewPubSubHandler(logger zerolog.Logger) *PubSubHandler {
	return &PubSubHandler{
		logger:     logger,
		clientPool: &sync.Map{},
	}
}

type PubSubHandler struct {
	logger     zerolog.Logger
	clientPool *sync.Map
}

func (handler *PubSubHandler) getClient(ctx context.Context, projectID, credential string) (*pubsub.Client, error) {
	key := _PubSubClientKey{
		ProjectID:  projectID,
		Credential: credential,
	}
	if v, ok := handler.clientPool.Load(key); ok {
		if client, ok := v.(*pubsub.Client); ok {
			return client, nil
		}
	}

	client, err := pubsub.NewClient(ctx, projectID, option.WithCredentialsJSON([]byte(credential)))
	if err != nil {
		return nil, fmt.Errorf("pubsub handler new client error: %v", err)
	}
	defer handler.clientPool.Store(key, client)

	return client, nil
}

func (handler *PubSubHandler) Do(ctx context.Context, endpointConfig *model.Endpoint) (map[string]string, interface{}, error) {
	handler.logger.Debug().Msg("start pubsub endpoint")
	defer handler.logger.Debug().Msg("complete pubsub endpoint")
	handler.logger.Debug().Interface("body", endpointConfig).Msgf("endpoint config")

	var (
		credential, projectID, topicID, body string
	)
	if endpointConfig.Credential != nil {
		credential = *endpointConfig.Credential
	}
	if endpointConfig.ProjectID != nil {
		projectID = *endpointConfig.ProjectID
	}
	if endpointConfig.TopicID != nil {
		topicID = *endpointConfig.TopicID
	}
	if endpointConfig.Body != nil {
		body = *endpointConfig.Body
	}

	client, err := handler.getClient(ctx, projectID, credential)
	if err != nil {
		return nil, nil, err
	}

	t := client.TopicInProject(topicID, projectID)
	serverID, err := t.Publish(ctx, &pubsub.Message{
		Data: []byte(body),
	}).Get(ctx)

	return nil, serverID, err
}
