package mongo_subscription

import (
	"context"

	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"github.com/th1enq/es-demo/config"
	"github.com/th1enq/es-demo/internal/domain"
	"github.com/th1enq/es-demo/internal/mappers"
	"github.com/th1enq/es-demo/internal/service"
	"github.com/th1enq/es-demo/pkg/es"
	"github.com/th1enq/es-demo/pkg/es/serializer"
	"go.uber.org/zap"
)

type MongoSubscription struct {
	log                *zap.Logger
	cfg                *config.Config
	bankAccountService *service.BankAccountService
	projection         es.Projection
	serializer         es.Serializer
	mongoRepository    domain.MongoRepository
	aggregateStore     es.AggregateStore
	eventBus           es.EventsBus
}

func NewBankAccountMongoSubscription(
	log *zap.Logger,
	cfg *config.Config,
	bankAccountService *service.BankAccountService,
	projection es.Projection,
	serializer es.Serializer,
	mongoRepository domain.MongoRepository,
	aggregateStore es.AggregateStore,
	eventBus es.EventsBus,
) *MongoSubscription {
	return &MongoSubscription{
		log:                log,
		cfg:                cfg,
		bankAccountService: bankAccountService,
		projection:         projection,
		serializer:         serializer,
		mongoRepository:    mongoRepository,
		aggregateStore:     aggregateStore,
		eventBus:           eventBus,
	}
}

func (s *MongoSubscription) ProcessMessagesErrGroup(ctx context.Context, r *kafka.Reader, workerID int) error {

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		m, err := r.FetchMessage(ctx)
		if err != nil {
			s.log.Warn("mongoSubscription.FetchMessage: %v", zap.Error(err))
			continue
		}

		// s.logProcessMessage(m, workerID)

		switch m.Topic {
		case es.GetTopicName(s.cfg.KafkaPublisherConfig.TopicPrefix, string(domain.BankAccountAggregateType)):
			s.handleBankAccountEvents(ctx, r, m)
		}
	}
}

func (s *MongoSubscription) handleBankAccountEvents(ctx context.Context, r *kafka.Reader, m kafka.Message) {

	var events []es.Event
	if err := serializer.Unmarshal(m.Value, &events); err != nil {
		s.log.Error("serializer.Unmarshal", zap.Error(err))
		// s.commitErrMessage(ctx, r, m)
		return
	}

	for _, event := range events {
		if err := s.handle(ctx, r, m, event); err != nil {
			return
		}
	}
	s.commitMessage(ctx, r, m)
}

func (s *MongoSubscription) handle(ctx context.Context, r *kafka.Reader, m kafka.Message, event es.Event) error {
	err := s.projection.When(ctx, event)
	if err != nil {
		s.log.Error("MongoSubscription When err", zap.Error(err))

		recreateErr := s.recreateProjection(ctx, event)
		if recreateErr != nil {
			return errors.Wrapf(recreateErr, "recreateProjection err: %v", err)
		}

		s.commitErrMessage(ctx, r, m)
		return errors.Wrapf(err, "When type: %s, aggregateID: %s", event.GetEventType(), event.GetAggregateID())
	}

	s.log.Info("MongoSubscription <<<commit>>> event: %s", zap.String("event", event.String()))
	return nil
}

func (s *MongoSubscription) recreateProjection(ctx context.Context, event es.Event) error {
	s.log.Warn("MongoSubscription recreating projection", zap.String("aggregateID", event.GetAggregateID()), zap.Uint64("version", event.GetVersion()), zap.Any("type", event.GetEventType()))

	err := s.mongoRepository.DeleteByAggregateID(ctx, event.GetAggregateID())
	if err != nil {
		s.log.Error("MongoSubscription DeleteByAggregateID", zap.Error(err))
		return errors.Wrapf(err, "When DeleteByAggregateID type: %s, aggregateID: %s", event.GetEventType(), event.GetAggregateID())
	}

	bankAccountAggregate := domain.NewBankAccountAggregate(event.GetAggregateID())
	err = s.aggregateStore.Load(ctx, bankAccountAggregate)
	if err != nil {
		s.log.Error("MongoSubscription aggregateStore.Load", zap.Error(err))
		return errors.Wrapf(err, "When aggregateStore.Load type: %s, aggregateID: %s", event.GetEventType(), event.GetAggregateID())
	}

	err = s.mongoRepository.Insert(ctx, mappers.BankAccountToMongoProjection(bankAccountAggregate))
	if err != nil {
		s.log.Error("MongoSubscription mongoRepository.Insert", zap.Error(err))
		return errors.Wrapf(err, "When mongoRepository.Insert type: %s, aggregateID: %s", event.GetEventType(), event.GetAggregateID())
	}

	s.log.Info("MongoSubscription <<<projection recreated commit>>> aggregateID: %s, version: %d", zap.String("aggregateID", event.GetAggregateID()), zap.Uint64("version", bankAccountAggregate.GetVersion()))
	return nil
}
