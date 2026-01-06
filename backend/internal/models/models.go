package models

import (
	"time"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type Transaction struct {
	ID             string    `json:"transaction_id" db:"transaction_id"`
	UserID         uuid.UUID `json:"user_id" db:"user_id"`
	Source         string    `json:"source" db:"source"`
	Account        string    `json:"account" db:"account"`
	Date           string    `json:"date" db:"date"`
	Amount         float64   `json:"amount" db:"amount"`
	Currency       string    `json:"currency" db:"currency"`
	Description    string    `json:"description" db:"description"`
	Direction      string    `json:"direction" db:"direction"`
	Merchant       *string   `json:"merchant" db:"merchant"`
	Category       *string   `json:"category" db:"category"`
	Subcategory    *string   `json:"subcategory" db:"subcategory"`
	Balance        *float64   `json:"balance" db:"balance"`
	IsTransfer     bool      `json:"is_transfer" db:"is_transfer"`
	IsFee          bool      `json:"is_fee" db:"is_fee"`
	IsTax          bool      `json:"is_tax" db:"is_tax"`
	Neutralized    bool      `json:"neutralized" db:"neutralized"`
	ProcessedAt    time.Time `json:"processed_at" db:"processed_at"`
}
