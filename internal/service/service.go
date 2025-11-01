package service

import (
	"github.com/th1enq/es-demo/internal/command"
	"github.com/th1enq/es-demo/internal/domain"
	"github.com/th1enq/es-demo/internal/query"
	"github.com/th1enq/es-demo/pkg/es"
	"go.uber.org/zap"
)

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
		aggregateStore,
		logger,
	)

	return &BankAccountService{
		Commands: bankAccountCommand,
		Query:    bankAccountQuery,
	}
}
