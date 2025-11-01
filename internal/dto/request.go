package dto

import (
	"time"

	"github.com/th1enq/es-demo/internal/domain"
)

type CreateBankAccountRequest struct {
	AggregateID string `json:"id" validate:"required,gte=0"`
	Email       string `json:"email" validate:"required,email"`
	FirstName   string `json:"first_name" validate:"required,gte=0"`
	LastName    string `json:"last_name" validate:"required,gte=0"`
	Balance     int64  `json:"balance" validate:"required,gte=0"`
	Status      string `json:"status"`
	Password    string `json:"password" validate:"required,min=6"` // Add password field
}

// Authentication DTOs
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int      `json:"expires_in"`
	User         UserInfo `json:"user"`
}

type RegisterRequest struct {
	ID             string `json:"id" validate:"required"`
	Email          string `json:"email" validate:"required,email"`
	FirstName      string `json:"first_name" validate:"required"`
	LastName       string `json:"last_name" validate:"required"`
	Password       string `json:"password" validate:"required,min=6"`
	InitialBalance int64  `json:"initial_balance" validate:"gte=0"`
}

type RegisterResponse struct {
	UserID  string `json:"user_id"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

type UserInfo struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
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
