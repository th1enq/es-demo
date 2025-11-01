package query

import (
	"github.com/th1enq/es-demo/pkg/es"
	"go.uber.org/zap"
)

type BankAccountQuery struct {
	GetBankAccountByID    GetBankAccountByID
	GetBankAccountByEmail GetBankAccountByEmail
	GetEventsHistory      *GetEventsHistoryQueryHandler
}

func NewBankAccountQuery(
	getBankAccountByID GetBankAccountByID,
	getBankAccountByEmail GetBankAccountByEmail,
	aggregateStore es.AggregateStore,
	log *zap.Logger,
) *BankAccountQuery {
	return &BankAccountQuery{
		GetBankAccountByID:    getBankAccountByID,
		GetBankAccountByEmail: getBankAccountByEmail,
		GetEventsHistory:      NewGetEventsHistoryQueryHandler(aggregateStore, log),
	}
}
