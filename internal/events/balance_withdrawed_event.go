package events

import "github.com/th1enq/es-demo/pkg/es"

const (
	BalanceWithdrawedEventTypeV1 es.EventType = "BALANCE_WITHDRAWED_V1"
)

type BalanceWithdrawedEventV1 struct {
	Amount    int64  `json:"amount"`
	PaymentID string `json:"payment_id"`
	Metadata  []byte `json:"-"`
}
