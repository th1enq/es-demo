package domain

import (
	"context"

	"github.com/Rhymond/go-money"
	"github.com/pkg/errors"
	bankAccountErrors "github.com/th1enq/es-demo/internal/errors"
	"github.com/th1enq/es-demo/internal/events"
	"github.com/th1enq/es-demo/pkg/es"
)

const (
	BankAccountAggregateType es.AggregateType = "BankAccount"
)

type BankAccountAggregate struct {
	*es.AggregateBase
	BankAccount *BankAccount
}

func NewBankAccountAggregate(id string) *BankAccountAggregate {
	if id == "" {
		return nil
	}

	bankAccountAggregate := &BankAccountAggregate{BankAccount: NewBankAccount(id)}
	aggregateBase := es.NewAggregateBase(bankAccountAggregate.When)
	aggregateBase.SetType(BankAccountAggregateType)
	aggregateBase.SetID(id)
	bankAccountAggregate.AggregateBase = aggregateBase
	return bankAccountAggregate
}

func (a *BankAccountAggregate) When(event any) error {

	switch evt := event.(type) {

	case *events.BankAccountCreatedEventV1:
		a.BankAccount.Email = evt.Email
		a.BankAccount.Balance = evt.Balance
		a.BankAccount.FirstName = evt.FirstName
		a.BankAccount.LastName = evt.LastName
		a.BankAccount.PasswordHash = evt.PasswordHash
		return nil

	case *events.BalanceDepositedEventV1:
		return a.BankAccount.Deposit(evt.Amount)

	case *events.BalanceWithdrawedEventV1:
		return a.BankAccount.Withdraw(evt.Amount)

	default:
		return errors.Wrapf(bankAccountErrors.ErrUnknownEventType, "event: %#v", event)
	}
}

func (a *BankAccountAggregate) CreateBankAccount(
	ctx context.Context,
	email, firstName, lastName string,
	amount int64,
	password string, // Add password parameter
) error {
	if amount < 0 {
		return errors.Wrapf(bankAccountErrors.ErrInvalidBalanceAmount, "amount: %d", amount)
	}

	// Create a temporary bank account to hash the password
	tempAccount := NewBankAccount(a.GetID())
	if err := tempAccount.HashPassword(password); err != nil {
		return errors.Wrap(err, "failed to hash password")
	}

	event := &events.BankAccountCreatedEventV1{
		Email:        email,
		FirstName:    firstName,
		LastName:     lastName,
		Balance:      money.New(amount, money.VND),
		PasswordHash: tempAccount.PasswordHash,
	}

	return a.Apply(event)
}

func (a *BankAccountAggregate) DepositBalance(ctx context.Context, amount int64, paymentID string) error {
	if amount <= 0 {
		return errors.Wrapf(bankAccountErrors.ErrInvalidBalanceAmount, "amount: %d", amount)
	}
	event := &events.BalanceDepositedEventV1{
		Amount:    amount,
		PaymentID: paymentID,
	}

	return a.Apply(event)
}

func (a *BankAccountAggregate) WithdrawBalance(ctx context.Context, amount int64, paymentID string) error {
	if amount <= 0 {
		return errors.Wrapf(bankAccountErrors.ErrInvalidBalanceAmount, "amount: %d", amount)
	}

	balance, err := money.New(a.BankAccount.Balance.Amount(), money.VND).Subtract(money.New(amount, money.VND))
	if err != nil {
		return errors.Wrapf(err, "Balance.Subtract amount: %d", amount)
	}

	if balance.IsNegative() {
		return errors.Wrapf(bankAccountErrors.ErrNotEnoughBalance, "amount: %d", amount)
	}

	event := &events.BalanceWithdrawedEventV1{
		Amount:    amount,
		PaymentID: paymentID,
	}

	return a.Apply(event)
}
