package service

import (
	"context"

	"github.com/th1enq/es-demo/internal/command"
	"github.com/th1enq/es-demo/internal/domain"
	"github.com/th1enq/es-demo/internal/dto"
	"github.com/th1enq/es-demo/internal/query"
	"github.com/th1enq/es-demo/pkg/es"
	"go.uber.org/zap"
)

// QueryService interface for querying data
type QueryService interface {
	GetBankAccountByID(ctx context.Context, id string) (*domain.BankAccount, error)
	GetBankAccountByEmail(ctx context.Context, email string) (*domain.BankAccount, error)
}

// CommandBus interface for executing commands
type CommandBus interface {
	CreateBankAccount(ctx context.Context, req dto.CreateBankAccountRequest) error
	DepositBalance(ctx context.Context, id string, amount int64, paymentID string) error
	WithdrawBalance(ctx context.Context, id string, amount int64, paymentID string) error
}

type BankAccountService struct {
	Commands *command.BankAccountCommand
	Query    *query.BankAccountQuery
}

func NewBankAccountService(
	logger *zap.Logger,
	aggregateStore es.AggregateStore,
	serializer es.Serializer,
	mongoRepository domain.MongoRepository,
) *BankAccountService {
	bankAccountCommand := command.NewBankAccountCommand(
		command.NewCreateBankAccountCmdHandler(aggregateStore, logger),
		command.NewDepositeBalanceCmdHandler(aggregateStore, logger),
		command.NewWithdrawBalanceCmdHandler(aggregateStore, logger),
	)

	bankAccountQuery := query.NewBankAccountQuery(
		query.NewGetBankAccountByIDQuery(
			mongoRepository,
			aggregateStore,
			logger,
		),
		query.NewGetBankAccountByEmailQuery(
			mongoRepository,
			logger,
		),
		aggregateStore,
		logger,
	)

	return &BankAccountService{
		Commands: bankAccountCommand,
		Query:    bankAccountQuery,
	}
}

// Implement QueryService interface
func (s *BankAccountService) GetBankAccountByID(ctx context.Context, id string) (*domain.BankAccount, error) {
	projection, err := s.Query.GetBankAccountByID.Handle(ctx, query.GetBankAccountByIDQuery{
		AggregateID: id,
	})
	if err != nil {
		return nil, err
	}

	// Convert projection to domain model
	return s.projectionToBankAccount(projection), nil
}

func (s *BankAccountService) GetBankAccountByEmail(ctx context.Context, email string) (*domain.BankAccount, error) {
	projection, err := s.Query.GetBankAccountByEmail.Handle(ctx, query.GetBankAccountByEmailQuery{
		Email: email,
	})
	if err != nil {
		return nil, err
	}

	// Convert projection to domain model
	return s.projectionToBankAccount(projection), nil
}

// Helper method to convert projection to domain model
func (s *BankAccountService) projectionToBankAccount(projection *domain.BankAccountMongoProjection) *domain.BankAccount {
	return &domain.BankAccount{
		AggregateID:  projection.AggregateID,
		Email:        projection.Email,
		FirstName:    projection.FirstName,
		LastName:     projection.LastName,
		PasswordHash: projection.PasswordHash,
		CreatedAt:    projection.CreatedAt,
		UpdatedAt:    projection.UpdatedAt,
		// Balance conversion from projection Balance to money.Money will be handled by the domain
		// For authentication purposes, we mainly need the other fields
	}
}

// Implement CommandBus interface
func (s *BankAccountService) CreateBankAccount(ctx context.Context, req dto.CreateBankAccountRequest) error {
	createCmd := command.CreateBankAccountCommand{
		AggregateID: req.AggregateID,
		Email:       req.Email,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Balance:     req.Balance,
		Password:    req.Password,
	}
	return s.Commands.CreateBankAccount.Handle(ctx, createCmd)
}

func (s *BankAccountService) DepositBalance(ctx context.Context, id string, amount int64, paymentID string) error {
	depositCmd := command.DepositeBalanceCommand{
		AggregateID: id,
		Amount:      amount,
		PaymentID:   paymentID,
	}
	return s.Commands.DepositeBalance.Handle(ctx, depositCmd)
}

func (s *BankAccountService) WithdrawBalance(ctx context.Context, id string, amount int64, paymentID string) error {
	withdrawCmd := command.WithdrawBalanceCommand{
		AggregateID: id,
		Amount:      amount,
		PaymentID:   paymentID,
	}
	return s.Commands.WithdrawBalance.Handle(ctx, withdrawCmd)
}
