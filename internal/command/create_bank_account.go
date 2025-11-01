package command

import (
	"context"

	"github.com/th1enq/es-demo/internal/domain"
	bankAccountErrors "github.com/th1enq/es-demo/internal/errors"
	"github.com/th1enq/es-demo/pkg/es"
	"go.uber.org/zap"
)

type CreateBankAccountCommand struct {
	AggregateID string `json:"id" validate:"required,gte=0"`
	Email       string `json:"email" validate:"required,gte=0,email"`
	FirstName   string `json:"first_name" validate:"required,gte=0"`
	LastName    string `json:"last_name" validate:"required,gte=0"`
	Balance     int64  `json:"balance" validate:"required,gte=0"`
	Password    string `json:"password" validate:"required,min=6"` // Add password field
}

type CreateBankAccount interface {
	Handle(ctx context.Context, cmd CreateBankAccountCommand) error
}

type createBankAccount struct {
	aggregateStore es.AggregateStore
	logger         *zap.Logger
}

func NewCreateBankAccountCmdHandler(
	aggregateStore es.AggregateStore,
	logger *zap.Logger,
) CreateBankAccount {
	return &createBankAccount{
		aggregateStore: aggregateStore,
		logger:         logger,
	}
}

func (c *createBankAccount) Handle(ctx context.Context, cmd CreateBankAccountCommand) error {
	c.logger.Info("Handling CreateBankAccountCommand", zap.String("id", cmd.AggregateID))
	exists, err := c.aggregateStore.Exists(ctx, cmd.AggregateID)
	if err != nil {
		return err
	}
	if exists {
		return bankAccountErrors.ErrBankAccountAlreadyExists
	}

	bankAccountAggregate := domain.NewBankAccountAggregate(cmd.AggregateID)
	err = bankAccountAggregate.CreateBankAccount(
		ctx,
		cmd.Email,
		cmd.FirstName,
		cmd.LastName,
		cmd.Balance,
		cmd.Password, // Add password parameter
	)
	if err != nil {
		return err
	}
	return c.aggregateStore.Save(ctx, bankAccountAggregate)
}
