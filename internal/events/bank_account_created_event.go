package events

import "github.com/Rhymond/go-money"

const (
	BankAccountCreatedEventTypeV1 = "BANK_ACCOUNT_CREATED_V1"
)

type BankAccountCreatedEventV1 struct {
	Email     string       `json:"email"`
	FirstName string       `json:"first_name"`
	LastName  string       `json:"last_name"`
	Balance   *money.Money `json:"balance"`
	Status    string       `json:"status"`
	Metadata  []byte       `json:"-"`
}
