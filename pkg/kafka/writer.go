package kafka

import (
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
	"go.uber.org/zap"
)

// NewWriter create new configured kafka writer
func NewWriter(brokers []string) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		MaxAttempts:  writerMaxAttempts,
		// ErrorLogger:  errLogger,
		Compression:  compress.Snappy,
		ReadTimeout:  writerReadTimeout,
		WriteTimeout: writerWriteTimeout,
		BatchTimeout: batchTimeout,
		BatchSize:    batchSize,
		Async:        false,
	}
}

// NewAsyncWriter create new configured kafka async writer
func NewAsyncWriter(brokers []string, log *zap.Logger) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		MaxAttempts:  writerMaxAttempts,
		// ErrorLogger:  errLogger,
		Compression:  compress.Snappy,
		ReadTimeout:  writerReadTimeout,
		WriteTimeout: writerWriteTimeout,
		Async:        true,
		Completion: func(messages []kafka.Message, err error) {
			if err != nil {
				log.Error("(kafka.AsyncWriter Error) topic: %s, partition: %v, offset: %v err: %v", zap.String("topic", messages[0].Topic), zap.Int("partition", messages[0].Partition), zap.Int64("offset", messages[0].Offset), zap.Error(err))
				return
			}
		},
	}
}

type AsyncWriterCallback func(messages []kafka.Message, logger *zap.Logger) error

// NewAsyncWriterWithCallback create new configured kafka async writer
func NewAsyncWriterWithCallback(brokers []string, logger *zap.Logger, cb AsyncWriterCallback) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		MaxAttempts:  writerMaxAttempts,
		// ErrorLogger:  errLogger,
		Compression:  compress.Snappy,
		ReadTimeout:  writerReadTimeout,
		WriteTimeout: writerWriteTimeout,
		Async:        true,
		Completion: func(messages []kafka.Message, err error) {
			if err != nil {
				logger.Error("(kafka.AsyncWriter Error) topic: %s, partition: %v, offset: %v err: %v", zap.String("topic", messages[0].Topic), zap.Int("partition", messages[0].Partition), zap.Int64("offset", messages[0].Offset), zap.Error(err))
				if err := cb(messages, logger); err != nil {
					logger.Error("(kafka.AsyncWriter Callback Error) err: %v", zap.Error(err))
					return
				}
				return
			}
		},
	}
}

// NewRequireNoneWriter create new configured kafka writer
func NewRequireNoneWriter(brokers []string, logger *zap.Logger) *kafka.Writer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireNone,
		MaxAttempts:  writerMaxAttempts,
		// ErrorLogger:  errLogger,
		Compression:  compress.Snappy,
		ReadTimeout:  writerRequireNoneReadTimeout,
		WriteTimeout: writerRequireNoneWriteTimeout,
		Async:        false,
		Completion: func(messages []kafka.Message, err error) {
			if err != nil {
				logger.Error("(kafka.Writer Error) topic: %s, partition: %v, offset: %v err: %v", zap.String("topic", messages[0].Topic), zap.Int("partition", messages[0].Partition), zap.Int64("offset", messages[0].Offset), zap.Error(err))
				return
			}
		},
	}
	return w
}
