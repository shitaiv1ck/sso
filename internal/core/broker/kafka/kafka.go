package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Producer interface {
	ProduceSync(ctx context.Context, rs ...*kgo.Record) kgo.ProduceResults

	NewRecord(topic string, msg any) (*kgo.Record, error)
	GetTimeout() time.Duration
}

type KafkaConn struct {
	*kgo.Client

	timeout time.Duration
}

func NewKafkaConn(ctx context.Context, config Config) (*KafkaConn, error) {
	seed := fmt.Sprintf("%s:%s", config.Host, config.Port)

	cl, err := kgo.NewClient(kgo.SeedBrokers(seed))
	if err != nil {
		return nil, err
	}

	if err := cl.Ping(ctx); err != nil {
		return nil, err
	}

	return &KafkaConn{
		Client:  cl,
		timeout: config.Timeout,
	}, nil
}

func (p *KafkaConn) NewRecord(topic string, msg any) (*kgo.Record, error) {
	value, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	record := &kgo.Record{Topic: topic, Value: value}

	return record, nil
}

func (p *KafkaConn) GetTimeout() time.Duration {
	return p.timeout
}
