package nats

import (
	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
)

var (
	_ _NatsClient = (*_NatsNormalClient)(nil)
	_ _NatsClient = (*_NatsStanClient)(nil)
)

type _NatsClient interface {
	SetFlushTime(t int64)
	FlushTime() int64
	Publish(subject string, data []byte) error
	Close() error
}

type _NatsStanClient struct {
	conn      stan.Conn
	flushTime int64
}

func (client *_NatsStanClient) Publish(subject string, data []byte) error {
	return client.conn.Publish(subject, data)
}

func (client *_NatsStanClient) Close() error {
	return client.conn.Close()
}

func (client *_NatsStanClient) SetFlushTime(t int64) {
	client.flushTime = t
}

func (client *_NatsStanClient) FlushTime() int64 {
	return client.flushTime
}

type _NatsNormalClient struct {
	conn      *nats.Conn
	flushTime int64
}

func (client *_NatsNormalClient) Publish(subject string, data []byte) error {
	return client.conn.Publish(subject, data)
}

func (client *_NatsNormalClient) Close() error {
	client.conn.Close()
	return nil
}

func (client *_NatsNormalClient) SetFlushTime(t int64) {
	client.flushTime = t
}

func (client *_NatsNormalClient) FlushTime() int64 {
	return client.flushTime
}
