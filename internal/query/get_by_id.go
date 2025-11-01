package query

import (
	"context"

	"github.com/pkg/errors"
	"github.com/th1enq/es-demo/internal/domain"
	bankAccountErrors "github.com/th1enq/es-demo/internal/errors"
	"github.com/th1enq/es-demo/internal/mappers"
	"github.com/th1enq/es-demo/pkg/es"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type GetBankAccountByIDQuery struct {
	AggregateID    string `json:"aggregate_id" validate:"required,gte=0"`
	FromEventStore bool   `json:"from_event_store"`
}

type GetBankAccountByID interface {
	Handle(ctx context.Context, query GetBankAccountByIDQuery) (*domain.BankAccountMongoProjection, error)
}

type getBankAccountByIDQuery struct {
	aggregateStore  es.AggregateStore
	mongoRepository domain.MongoRepository
	logger          *zap.Logger
}

func NewGetBankAccountByIDQuery(
	bankAccountRepo domain.MongoRepository,
	aggregateStore es.AggregateStore,
	logger *zap.Logger,
) GetBankAccountByID {
	return &getBankAccountByIDQuery{
		mongoRepository: bankAccountRepo,
		aggregateStore:  aggregateStore,
		logger:          logger,
	}
}

func (q *getBankAccountByIDQuery) Handle(ctx context.Context, query GetBankAccountByIDQuery) (*domain.BankAccountMongoProjection, error) {
	q.logger.Info("query", zap.Any("query", query))
	if query.FromEventStore {
		return q.loadFromAggregateStore(ctx, query)
	}

	projection, err := q.mongoRepository.GetByAggregateID(ctx, query.AggregateID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			bankAccountAggregate := domain.NewBankAccountAggregate(query.AggregateID)
			if err = q.aggregateStore.Load(ctx, bankAccountAggregate); err != nil {
				return nil, err
			}
			if bankAccountAggregate.GetVersion() == 0 {
				return nil, errors.Wrapf(bankAccountErrors.ErrBankAccountNotFound, "id: %s", query.AggregateID)
			}

			mongoProjection := mappers.BankAccountToMongoProjection(bankAccountAggregate)
			err = q.mongoRepository.Upsert(ctx, mongoProjection)
			if err != nil {
				q.logger.Error("MongoDB upsert failed", zap.String("AggregateID", query.AggregateID), zap.Error(err))
			}
			return mongoProjection, nil

		}
		return nil, err
	}

	return projection, nil
}

func (q *getBankAccountByIDQuery) loadFromAggregateStore(ctx context.Context, query GetBankAccountByIDQuery) (*domain.BankAccountMongoProjection, error) {

	bankAccountAggregate := domain.NewBankAccountAggregate(query.AggregateID)
	if err := q.aggregateStore.Load(ctx, bankAccountAggregate); err != nil {
		return nil, err
	}
	if bankAccountAggregate.GetVersion() == 0 {
		return nil, errors.Wrapf(bankAccountErrors.ErrBankAccountNotFound, "id: %s", query.AggregateID)
	}

	return mappers.BankAccountToMongoProjection(bankAccountAggregate), nil
}
