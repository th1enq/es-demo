package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Producer interface {
	PublishMessage(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type producer struct {
	log     *zap.Logger
	brokers []string
	w       *kafka.Writer
}

// NewProducer create new kafka producer
func NewProducer(log *zap.Logger, brokers []string) *producer {
	return &producer{log: log, brokers: brokers, w: NewWriter(brokers)}
}

// NewAsyncProducer create new kafka producer
func NewAsyncProducer(log *zap.Logger, brokers []string) *producer {
	return &producer{log: log, brokers: brokers, w: NewAsyncWriter(brokers, log)}
}

// NewAsyncProducerWithCallback create new kafka producer with callback for delete invalid projection
func NewAsyncProducerWithCallback(log *zap.Logger, brokers []string, cb AsyncWriterCallback) *producer {
	return &producer{log: log, brokers: brokers, w: NewAsyncWriterWithCallback(brokers, log, cb)}
}

// NewRequireNoneProducer create new fire and forget kafka producer
func NewRequireNoneProducer(log *zap.Logger, brokers []string) *producer {
	return &producer{log: log, brokers: brokers, w: NewRequireNoneWriter(brokers, log)}
}

func (p *producer) PublishMessage(ctx context.Context, msgs ...kafka.Message) error {

	if err := p.w.WriteMessages(ctx, msgs...); err != nil {
		return err
	}
	return nil
}

func (p *producer) Close() error {
	return p.w.Close()
}
