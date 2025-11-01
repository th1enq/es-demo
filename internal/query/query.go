package query

import (
	"github.com/th1enq/es-demo/pkg/es"
	"go.uber.org/zap"
)

type BankAccountQuery struct {
	GetBankAccountByID GetBankAccountByID
	GetEventsHistory   *GetEventsHistoryQueryHandler
}

func NewBankAccountQuery(
	getBankAccountByID GetBankAccountByID,
	aggregateStore es.AggregateStore,
	log *zap.Logger,
) *BankAccountQuery {
	return &BankAccountQuery{
		GetBankAccountByID: getBankAccountByID,
		GetEventsHistory:   NewGetEventsHistoryQueryHandler(aggregateStore, log),
	}
}
