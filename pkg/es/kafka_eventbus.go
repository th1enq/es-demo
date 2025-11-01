package es

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"github.com/th1enq/es-demo/pkg/es/serializer"
	kafkaClient "github.com/th1enq/es-demo/pkg/kafka"
)

// KafkaEventsBusConfig kafka eventbus config.
type KafkaEventsBusConfig struct {
	Topic             string `mapstructure:"topic" validate:"required"`
	TopicPrefix       string `mapstructure:"topicPrefix" validate:"required"`
	Partitions        int    `mapstructure:"partitions" validate:"required,gte=0"`
	ReplicationFactor int    `mapstructure:"replicationFactor" validate:"required,gte=0"`
	Headers           []kafka.Header
}

type kafkaEventsBus struct {
	producer kafkaClient.Producer
	cfg      KafkaEventsBusConfig
}

// NewKafkaEventsBus kafkaEventsBus constructor.
func NewKafkaEventsBus(producer kafkaClient.Producer, cfg KafkaEventsBusConfig) *kafkaEventsBus {
	return &kafkaEventsBus{producer: producer, cfg: cfg}
}

// ProcessEvents serialize to json and publish es.Event's to the kafka topic.
func (e *kafkaEventsBus) ProcessEvents(ctx context.Context, events []Event) error {

	eventsBytes, err := serializer.Marshal(events)
	if err != nil {
		return errors.Wrap(err, "serializer.Marshal")
	}

	return e.producer.PublishMessage(ctx, kafka.Message{
		Topic:   GetTopicName(e.cfg.TopicPrefix, string(events[0].GetAggregateType())),
		Value:   eventsBytes,
		Headers: nil,
		Time:    time.Now().UTC(),
	})
}

func GetTopicName(eventStorePrefix string, aggregateType string) string {
	return fmt.Sprintf("%s_%s", eventStorePrefix, aggregateType)
}

func GetKafkaAggregateTypeTopic(cfg KafkaEventsBusConfig, aggregateType string) kafka.TopicConfig {
	return kafka.TopicConfig{
		Topic:             GetTopicName(cfg.TopicPrefix, aggregateType),
		NumPartitions:     cfg.Partitions,
		ReplicationFactor: cfg.ReplicationFactor,
	}
}
