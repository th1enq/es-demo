package errors

import "errors"

var (
	ErrUnknownEventType         = errors.New("unknown event type")
	ErrInvalidBalanceAmount     = errors.New("invalid amount")
	ErrNotEnoughBalance         = errors.New("balance has not enough balance")
	ErrBankAccountNotFound      = errors.New("bank account not found")
	ErrBankAccountAlreadyExists = errors.New("bank account with given id already exists")

	// Authentication errors
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrAccountInactive    = errors.New("account is inactive")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrUnauthorized       = errors.New("unauthorized access")
	ErrForbidden          = errors.New("forbidden access")
)
