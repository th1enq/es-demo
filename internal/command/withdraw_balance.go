package command

import (
	"context"

	"github.com/th1enq/es-demo/internal/domain"
	"github.com/th1enq/es-demo/pkg/es"
	"go.uber.org/zap"
)

type WithdrawBalanceCommand struct {
	AggregateID string `json:"aggregate_id" validate:"required,gte=0"`
	Amount      int64  `json:"amount" validate:"required,gt=0"`
	PaymentID   string `json:"payment_id" validate:"required,gte=0"`
}

type WithdrawBalance interface {
	Handle(ctx context.Context, cmd WithdrawBalanceCommand) error
}

type withdrawBalanceCmdHandler struct {
	aggregateStore es.AggregateStore
	logger         *zap.Logger
}

func NewWithdrawBalanceCmdHandler(
	aggregateStore es.AggregateStore,
	logger *zap.Logger,
) WithdrawBalance {
	return &withdrawBalanceCmdHandler{
		aggregateStore: aggregateStore,
		logger:         logger,
	}
}

func (w *withdrawBalanceCmdHandler) Handle(ctx context.Context, cmd WithdrawBalanceCommand) error {
	w.logger.Info("Handling WithdrawBalanceCommand", zap.String("id", cmd.AggregateID))
	NewBankAccountAggregate := domain.NewBankAccountAggregate(cmd.AggregateID)
	err := w.aggregateStore.Load(ctx, NewBankAccountAggregate)

	if err != nil {
		return err
	}

	if err := NewBankAccountAggregate.WithdrawBalance(
		ctx,
		cmd.Amount,
		cmd.PaymentID,
	); err != nil {
		return err
	}

	return w.aggregateStore.Save(ctx, NewBankAccountAggregate)
}
