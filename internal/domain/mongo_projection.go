package domain

import "time"

type BankAccountMongoProjection struct {
	ID           string    `json:"id" bson:"_id,omitempty"`
	AggregateID  string    `json:"aggregate_id" bson:"aggregate_id,omitempty"`
	Version      uint64    `json:"version" bson:"version,omitempty"`
	Email        string    `json:"email" bson:"email,omitempty"`
	FirstName    string    `json:"first_name" bson:"first_name,omitempty"`
	LastName     string    `json:"last_name" bson:"last_name,omitempty"`
	Balance      Balance   `json:"balance" bson:"balance,omitempty"`
	PasswordHash string    `json:"-" bson:"password_hash,omitempty"` // Don't expose in JSON
	CreatedAt    time.Time `json:"created_at" bson:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at" bson:"updated_at,omitempty"`
}

// CheckPassword verifies the password against the stored hash
func (p *BankAccountMongoProjection) CheckPassword(password string) error {
	account := &BankAccount{PasswordHash: p.PasswordHash}
	return account.CheckPassword(password)
}
