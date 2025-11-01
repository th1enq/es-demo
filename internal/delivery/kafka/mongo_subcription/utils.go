package mongo_subscription

import (
	"context"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

func (s *MongoSubscription) commitMessage(ctx context.Context, r *kafka.Reader, m kafka.Message) {
	if err := r.CommitMessages(ctx, m); err != nil {
		s.log.Error("(mongoSubscription) [CommitMessages] err: %v", zap.Error(err))
		return
	}
	// s.log.KafkaLogCommittedMessage(m.Topic, m.Partition, m.Offset)
}

func (s *MongoSubscription) commitErrMessage(ctx context.Context, r *kafka.Reader, m kafka.Message) {
	if err := r.CommitMessages(ctx, m); err != nil {
		s.log.Error("(mongoSubscription) [CommitMessages] err: %v", zap.Error(err))
		return
	}
	// s.log.KafkaLogCommittedMessage(m.Topic, m.Partition, m.Offset)
}

func (s *MongoSubscription) logProcessMessage(m kafka.Message, workerID int) {
	// s.log.KafkaProcessMessage(m.Topic, m.Partition, m.Value, workerID, m.Offset, m.Time)
}
