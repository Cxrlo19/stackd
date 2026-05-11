package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID               uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey"`
	Name             string         `json:"name" gorm:"not null"`
	Email            string         `json:"email" gorm:"uniqueIndex;not null"`
	Password         string         `json:"-" gorm:"not null"`
	IsVerified       bool           `json:"is_verified" gorm:"default:false"`
	StripeCustomerID *string        `json:"stripe_customer_id,omitempty"`
	Plan             string         `json:"plan" gorm:"default:free"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
}

type BankAccount struct {
	ID               uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	UserID           uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	PlaidAccountID   string     `json:"-" gorm:"not null"`
	PlaidAccessToken string     `json:"-" gorm:"not null"`
	InstitutionName  string     `json:"institution_name"`
	AccountName      string     `json:"account_name"`
	AccountType      string     `json:"account_type"`
	Balance          float64    `json:"balance"`
	Currency         string     `json:"currency" gorm:"default:USD"`
	LastSynced       *time.Time `json:"last_synced"`
	CreatedAt        time.Time  `json:"created_at"`
}

type Transaction struct {
	ID                 uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	UserID             uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	AccountID          uuid.UUID `json:"account_id" gorm:"type:uuid;not null"`
	PlaidTransactionID *string   `json:"-"`
	Name               string    `json:"name" gorm:"not null"`
	Amount             float64   `json:"amount" gorm:"not null"`
	Category           string    `json:"category"`
	Date               time.Time `json:"date"`
	IsRecurring        bool      `json:"is_recurring" gorm:"default:false"`
	Notes              *string   `json:"notes,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
}

type Budget struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	Category  string    `json:"category" gorm:"not null"`
	Amount    float64   `json:"amount" gorm:"not null"`
	Period    string    `json:"period" gorm:"default:monthly"`
	CreatedAt time.Time `json:"created_at"`
}

type Insight struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	Content   string    `json:"content" gorm:"not null"`
	Type      string    `json:"type" gorm:"default:weekly"`
	CreatedAt time.Time `json:"created_at"`
}
