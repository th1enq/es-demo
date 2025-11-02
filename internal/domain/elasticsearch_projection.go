package domain

import (
	"encoding/json"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/th1enq/es-demo/internal/events"
)

// BankAccountElasticsearchProjection represents bank account data in Elasticsearch
type BankAccountElasticsearchProjection struct {
	AggregateID string             `json:"aggregateId"`
	Email       string             `json:"email"`
	FirstName   string             `json:"firstName"`
	LastName    string             `json:"lastName"`
	Balance     *BalanceProjection `json:"balance"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
	Version     uint64             `json:"version"`
	// Additional fields for Elasticsearch analytics
	TotalDeposits    int64     `json:"totalDeposits"`
	TotalWithdrawals int64     `json:"totalWithdrawals"`
	TransactionCount int       `json:"transactionCount"`
	Status           string    `json:"status"` // active, inactive, frozen
	LastActivity     time.Time `json:"lastActivity"`
}

// BalanceProjection for Elasticsearch
type BalanceProjection struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

// NewBankAccountElasticsearchProjection creates a new Elasticsearch projection
func NewBankAccountElasticsearchProjection(aggregateID string) *BankAccountElasticsearchProjection {
	return &BankAccountElasticsearchProjection{
		AggregateID:      aggregateID,
		Balance:          &BalanceProjection{Amount: 0, Currency: "USD"},
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		Version:          0,
		TotalDeposits:    0,
		TotalWithdrawals: 0,
		TransactionCount: 0,
		Status:           "active",
		LastActivity:     time.Now(),
	}
}

// ToJSON converts the projection to JSON string
func (p *BankAccountElasticsearchProjection) ToJSON() (string, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON populates the projection from JSON string
func (p *BankAccountElasticsearchProjection) FromJSON(jsonData string) error {
	return json.Unmarshal([]byte(jsonData), p)
}

// GetBalance returns the balance as money.Money
func (p *BankAccountElasticsearchProjection) GetBalance() *money.Money {
	if p.Balance == nil {
		p.Balance = &BalanceProjection{Amount: 0, Currency: "USD"}
	}
	return money.New(p.Balance.Amount, p.Balance.Currency)
}

// SetBalance sets the balance from money.Money
func (p *BankAccountElasticsearchProjection) SetBalance(balance *money.Money) {
	if balance == nil {
		balance = money.New(0, "USD")
	}
	if p.Balance == nil {
		p.Balance = &BalanceProjection{}
	}
	p.Balance.Amount = balance.Amount()
	p.Balance.Currency = balance.Currency().Code
	p.UpdatedAt = time.Now()
	p.LastActivity = time.Now()
}

// When methods to handle events for replay

// When BankAccountCreatedEventV1 is applied
func (p *BankAccountElasticsearchProjection) WhenBankAccountCreated(event events.BankAccountCreatedEventV1, aggregateID string, version uint64, timestamp time.Time) {
	p.AggregateID = aggregateID
	p.Email = event.Email
	p.FirstName = event.FirstName
	p.LastName = event.LastName
	p.SetBalance(event.Balance)
	p.CreatedAt = timestamp
	p.UpdatedAt = timestamp
	p.Version = version
	p.Status = "active"
	p.LastActivity = timestamp
}

// When BalanceDepositedEventV1 is applied
func (p *BankAccountElasticsearchProjection) WhenBalanceDeposited(event events.BalanceDepositedEventV1, version uint64, timestamp time.Time) {
	currentBalance := p.GetBalance()
	depositAmount := money.New(event.Amount, "USD")
	newBalance, _ := currentBalance.Add(depositAmount)
	p.SetBalance(newBalance)
	p.TotalDeposits += event.Amount
	p.TransactionCount++
	p.Version = version
	p.UpdatedAt = timestamp
	p.LastActivity = timestamp
}

// When BalanceWithdrawedEventV1 is applied
func (p *BankAccountElasticsearchProjection) WhenBalanceWithdrawn(event events.BalanceWithdrawedEventV1, version uint64, timestamp time.Time) {
	currentBalance := p.GetBalance()
	withdrawAmount := money.New(event.Amount, "USD")
	newBalance, _ := currentBalance.Subtract(withdrawAmount)
	p.SetBalance(newBalance)
	p.TotalWithdrawals += event.Amount
	p.TransactionCount++
	p.Version = version
	p.UpdatedAt = timestamp
	p.LastActivity = timestamp
}

// GetFullName returns the full name
func (p *BankAccountElasticsearchProjection) GetFullName() string {
	return p.FirstName + " " + p.LastName
}

// IsActive checks if the account is active
func (p *BankAccountElasticsearchProjection) IsActive() bool {
	return p.Status == "active"
}

// GetNetFlow returns the net flow (deposits - withdrawals)
func (p *BankAccountElasticsearchProjection) GetNetFlow() int64 {
	return p.TotalDeposits - p.TotalWithdrawals
}

// UpdateStatus updates the account status
func (p *BankAccountElasticsearchProjection) UpdateStatus(status string) {
	p.Status = status
	p.UpdatedAt = time.Now()
	p.LastActivity = time.Now()
}

// Repository interface for Elasticsearch operations
type ElasticsearchRepository interface {
	CreateIndex(indexName string) error
	IndexDocument(indexName, documentID string, document interface{}) error
	GetDocument(indexName, documentID string) (*BankAccountElasticsearchProjection, error)
	UpdateDocument(indexName, documentID string, document interface{}) error
	DeleteDocument(indexName, documentID string) error
	Search(indexName string, query map[string]interface{}) ([]*BankAccountElasticsearchProjection, error)
	DeleteIndex(indexName string) error
	BulkIndex(indexName string, documents map[string]interface{}) error
}
