package projection

import (
	"context"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/pkg/errors"
	"github.com/th1enq/es-demo/internal/domain"
	bankAccountErrors "github.com/th1enq/es-demo/internal/errors"
	"github.com/th1enq/es-demo/internal/events"
	"github.com/th1enq/es-demo/pkg/es"
	"go.uber.org/zap"
)

type bankAccountMongoProjection struct {
	serializer      es.Serializer
	mongoRepository domain.MongoRepository
	logger          *zap.Logger
}

func NewBankAccountMongoProjection(
	serializer es.Serializer,
	mongoRepository domain.MongoRepository,
	logger *zap.Logger,
) *bankAccountMongoProjection {
	return &bankAccountMongoProjection{
		serializer:      serializer,
		mongoRepository: mongoRepository,
		logger:          logger,
	}
}

func (b *bankAccountMongoProjection) When(ctx context.Context, esEvent es.Event) error {
	deserializedEvent, err := b.serializer.DeserializeEvent(esEvent)

	if err != nil {
		return errors.Wrapf(err, "serializer.DeserializeEvent aggregateID: %s, type: %s", esEvent.GetAggregateID(), esEvent.GetEventType())
	}

	switch event := deserializedEvent.(type) {
	case *events.BalanceWithdrawedEventV1:
		return b.onBankAccountBalanceWithdrawed(ctx, esEvent, event)
	case *events.BalanceDepositedEventV1:
		return b.onBankAccountBalanceDeposited(ctx, esEvent, event)
	case *events.BankAccountCreatedEventV1:
		return b.onBankAccountCreated(ctx, esEvent, event)
	default:
		return errors.Wrapf(bankAccountErrors.ErrUnknownEventType, "esEvent: %s", esEvent.String())
	}
}

func (b *bankAccountMongoProjection) onBankAccountCreated(ctx context.Context, esEvent es.Event, event *events.BankAccountCreatedEventV1) error {
	b.logger.Info("Bank Account Create", zap.String("aggregate ID", esEvent.EventID))

	if esEvent.GetVersion() != 1 {
		return errors.Wrapf(es.ErrInvalidEventVersion, "type: %s, version: %d", esEvent.GetEventType(), esEvent.GetVersion())
	}

	projection := &domain.BankAccountMongoProjection{
		AggregateID: esEvent.GetAggregateID(),
		Version:     esEvent.GetVersion(),
		Email:       event.Email,
		FirstName:   event.FirstName,
		LastName:    event.LastName,
		Balance: domain.Balance{
			Amount:   event.Balance.AsMajorUnits(),
			Currency: event.Balance.Currency().Code,
		},
		PasswordHash: event.PasswordHash,
		UpdatedAt:    time.Now().UTC(),
		CreatedAt:    time.Now().UTC(),
	}

	err := b.mongoRepository.Insert(ctx, projection)
	if err != nil {
		return errors.Wrapf(err, "[onBankAccountCreated] mongoRepository.Insert aggregateID: %s", esEvent.GetAggregateID())
	}
	b.logger.Info("Bank Account Created projection", zap.Any("projection", projection))
	return nil
}

func (b *bankAccountMongoProjection) onBankAccountBalanceDeposited(ctx context.Context, esEvent es.Event, event *events.BalanceDepositedEventV1) error {
	b.logger.Info("Bank Account Deposit", zap.String("aggregate ID", esEvent.EventID))
	if err := b.mongoRepository.UpdateConcurrently(
		ctx,
		esEvent.GetAggregateID(),
		func(projection *domain.BankAccountMongoProjection) *domain.BankAccountMongoProjection {
			projection.Balance.Amount += float64(money.New(event.Amount, money.VND).Amount())
			projection.Version = esEvent.Version
			return projection
		},
		esEvent.GetVersion()-1,
	); err != nil {
		return errors.Wrapf(err, "[onBalanceDeposited] mongoRepository.UpdateConcurrently aggregateID: %s", esEvent.GetAggregateID())
	}
	b.logger.Info("Balance Deposited", zap.Any("event type", esEvent.GetEventType()), zap.String("aggregate id", esEvent.GetAggregateID()), zap.Uint64("version", esEvent.GetVersion()))
	return nil
}

func (b *bankAccountMongoProjection) onBankAccountBalanceWithdrawed(ctx context.Context, esEvent es.Event, event *events.BalanceWithdrawedEventV1) error {
	b.logger.Info("Bank Account Withdraw", zap.String("aggregate ID", esEvent.EventID))
	if err := b.mongoRepository.UpdateConcurrently(
		ctx,
		esEvent.GetAggregateID(),
		func(projection *domain.BankAccountMongoProjection) *domain.BankAccountMongoProjection {
			projection.Balance.Amount -= float64(money.New(event.Amount, money.VND).Amount())
			projection.Version = esEvent.Version
			return projection
		},
		esEvent.GetVersion()-1,
	); err != nil {
		return errors.Wrapf(err, "[onBalanceWithdrawed] mongoRepository.UpdateConcurrently aggregateID: %s", esEvent.GetAggregateID())
	}
	b.logger.Info("Balance Withdrawed", zap.Any("event type", esEvent.GetEventType()), zap.String("aggregate id", esEvent.GetAggregateID()), zap.Uint64("version", esEvent.GetVersion()))
	return nil
}
