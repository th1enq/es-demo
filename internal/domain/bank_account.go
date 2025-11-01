package domain

import (
	"time"

	"github.com/Rhymond/go-money"
	"golang.org/x/crypto/bcrypt"
)

type BankAccount struct {
	AggregateID string       `json:"aggregate_id"`
	Email       string       `json:"email"`
	FirstName   string       `json:"first_name"`
	LastName    string       `json:"last_name"`
	Balance     *money.Money `json:"balance"`
	Status      string       `json:"status"`
	// Authentication fields
	PasswordHash string    `json:"-"` // Don't expose in JSON
	Role         string    `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func NewBankAccount(
	id string,
) *BankAccount {
	return &BankAccount{
		AggregateID: id,
		Balance:     money.New(0, "USD"),
		Role:        "user", // Default role
		IsActive:    true,   // Default active
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// HashPassword hashes the password for authentication
func (b *BankAccount) HashPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	b.PasswordHash = string(hashedPassword)
	b.UpdatedAt = time.Now()
	return nil
}

// CheckPassword verifies the password
func (b *BankAccount) CheckPassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(b.PasswordHash), []byte(password))
}

func (b *BankAccount) Deposit(amount int64) error {
	result, err := b.Balance.Add(money.New(amount, money.USD))
	if err != nil {
		return err
	}
	b.Balance = result
	return nil
}

func (b *BankAccount) Withdraw(amount int64) error {
	result, err := b.Balance.Subtract(money.New(amount, money.USD))
	if err != nil {
		return err
	}
	b.Balance = result
	return nil
}
