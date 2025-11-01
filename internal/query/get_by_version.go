package query

import (
	"context"

	"github.com/pkg/errors"
	"github.com/th1enq/es-demo/internal/domain"
	bankAccountErrors "github.com/th1enq/es-demo/internal/errors"
	"github.com/th1enq/es-demo/internal/mappers"
	"github.com/th1enq/es-demo/pkg/es"
	"go.uber.org/zap"
)

type GetBankAccountByVersionQuery struct {
	AggregateID string `json:"aggregate_id" validate:"required,gte=0"`
	Version     uint64 `json:"version" validate:"required,gte=1"`
}

type GetBankAccountByVersion interface {
	Handle(ctx context.Context, query GetBankAccountByVersionQuery) (*domain.BankAccountMongoProjection, error)
}

type getBankAccountByVersionQuery struct {
	aggregateStore es.AggregateStore
	logger         *zap.Logger
}

func NewGetBankAccountByVersionQuery(
	aggregateStore es.AggregateStore,
	logger *zap.Logger,
) GetBankAccountByVersion {
	return &getBankAccountByVersionQuery{
		aggregateStore: aggregateStore,
		logger:         logger,
	}
}

func (q *getBankAccountByVersionQuery) Handle(ctx context.Context, query GetBankAccountByVersionQuery) (*domain.BankAccountMongoProjection, error) {
	q.logger.Info("GetBankAccountByVersion query",
		zap.String("aggregate_id", query.AggregateID),
		zap.Uint64("version", query.Version))

	// Create a new aggregate instance
	bankAccountAggregate := domain.NewBankAccountAggregate(query.AggregateID)

	// Load aggregate state at specific version using LoadByVersion
	if err := q.aggregateStore.LoadByVersion(ctx, bankAccountAggregate, query.Version); err != nil {
		q.logger.Error("Failed to load aggregate by version",
			zap.String("aggregate_id", query.AggregateID),
			zap.Uint64("version", query.Version),
			zap.Error(err))
		return nil, errors.Wrapf(err, "failed to load aggregate %s at version %d", query.AggregateID, query.Version)
	}

	// Check if aggregate exists (version should be > 0 after loading)
	if bankAccountAggregate.GetVersion() == 0 {
		return nil, errors.Wrapf(bankAccountErrors.ErrBankAccountNotFound,
			"aggregate_id: %s, version: %d", query.AggregateID, query.Version)
	}

	// Check if requested version exists
	if bankAccountAggregate.GetVersion() < query.Version {
		return nil, errors.Wrapf(bankAccountErrors.ErrBankAccountNotFound,
			"version %d not found for aggregate %s (current version: %d)",
			query.Version, query.AggregateID, bankAccountAggregate.GetVersion())
	}

	// Convert aggregate to mongo projection for response
	mongoProjection := mappers.BankAccountToMongoProjection(bankAccountAggregate)

	q.logger.Info("Successfully loaded bank account by version",
		zap.String("aggregate_id", query.AggregateID),
		zap.Uint64("requested_version", query.Version),
		zap.Uint64("loaded_version", bankAccountAggregate.GetVersion()))

	return mongoProjection, nil
}
