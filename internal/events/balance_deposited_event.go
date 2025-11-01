package events

import "github.com/th1enq/es-demo/pkg/es"

const (
	BalancedDepositedEventTypeV1 es.EventType = "BALANCE_DEPOSITED_V1"
)

type BalanceDepositedEventV1 struct {
	Amount    int64  `json:"amount"`
	PaymentID string `json:"payment_id"`
	Metadata  []byte `json:"-"`
}
