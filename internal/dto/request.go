package dto

import (
	"time"

	"github.com/th1enq/es-demo/internal/domain"
)

type CreateBankAccountRequest struct {
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"first_name" validate:"required,gte=0"`
	LastName  string `json:"last_name" validate:"required,gte=0"`
	Balance   int64  `json:"balance" validate:"required,gte=0"`
	Status    string `json:"status"`
	Password  string `json:"password" validate:"required,min=6"` // Add password field
}

type HttpBankAccountResponse struct {
	AggregateID string         `json:"aggregateID" bson:"aggregateID,omitempty"`
	Email       string         `json:"email" bson:"email,omitempty"`
	FirstName   string         `json:"firstName" bson:"firstName,omitempty"`
	LastName    string         `json:"lastName" bson:"lastName,omitempty"`
	Balance     domain.Balance `json:"balance" bson:"balance"`
	Status      string         `json:"status" bson:"status,omitempty"`
}

type RollbackRequest struct {
	Version uint64 `json:"version" validate:"required,gte=1"`
}

type EventResponse struct {
	EventID       string      `json:"event_id"`
	AggregateID   string      `json:"aggregate_id"`
	EventType     string      `json:"event_type"`
	AggregateType string      `json:"aggregate_type"`
	Version       uint64      `json:"version"`
	Data          interface{} `json:"data"`
	Metadata      interface{} `json:"metadata,omitempty"`
	Timestamp     time.Time   `json:"timestamp"`
}

type EventsHistoryResponse struct {
	AggregateID string          `json:"aggregate_id"`
	TotalEvents int             `json:"total_events"`
	Events      []EventResponse `json:"events"`
}
