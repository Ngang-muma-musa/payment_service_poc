package domain

import "context"

type Repository interface {
	Create(ctx context.Context, payment *Payment) error
	UpdateStatus(ctx context.Context, id string, status PaymentStatus) error
	FindByID(ctx context.Context, id string) (*Payment, error)
}
