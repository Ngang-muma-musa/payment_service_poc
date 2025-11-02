package domain

import "time"

type PaymentStatus string

const (
	StatusQueued     PaymentStatus = "queued"
	StatusProcessing PaymentStatus = "processing"
	StatusCompleted  PaymentStatus = "completed"
	StatusFailed     PaymentStatus = "failed"
)

type Payment struct {
	ID        string
	UserID    string        `json:"user_id"`
	Amount    float64       `json:"amount"`
	Currency  string        `json:"currency"`
	Status    PaymentStatus `json:"status"`
	CreatedAt time.Time     `json:"initiated_at"`
	UpdatedAt time.Time
}
