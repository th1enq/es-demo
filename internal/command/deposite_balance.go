package command

import (
	"context"

	"github.com/th1enq/es-demo/internal/domain"
	"github.com/th1enq/es-demo/pkg/es"
	"go.uber.org/zap"
)

type DepositeBalanceCommand struct {
	AggregateID string `json:"aggregate_id" validate:"required,gte=0"`
	Amount      int64  `json:"amount" validate:"required,gt=0"`
	PaymentID   string `json:"payment_id" validate:"required,gte=0"`
}

type DepositeBalance interface {
	Handle(ctx context.Context, cmd DepositeBalanceCommand) error
}

type depositeBalanceCmdHandler struct {
	aggregateStore es.AggregateStore
	logger         *zap.Logger
}

func NewDepositeBalanceCmdHandler(
	aggregateStore es.AggregateStore,
	logger *zap.Logger,
) DepositeBalance {
	return &depositeBalanceCmdHandler{
		aggregateStore: aggregateStore,
		logger:         logger,
	}
}

func (d *depositeBalanceCmdHandler) Handle(ctx context.Context, cmd DepositeBalanceCommand) error {
	d.logger.Info("Handling DepositeBalanceCommand", zap.String("id", cmd.AggregateID))
	bankAccoutAggregate := domain.NewBankAccountAggregate(cmd.AggregateID)
	err := d.aggregateStore.Load(ctx, bankAccoutAggregate)
	if err != nil {
		return err
	}

	if err := bankAccoutAggregate.DepositBalance(
		ctx,
		cmd.Amount,
		cmd.PaymentID,
	); err != nil {
		return err
	}

	return d.aggregateStore.Save(ctx, bankAccoutAggregate)
}
