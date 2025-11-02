package app

import (
	"context"
	"paymentservice/internal/domain"
)

type QueuePort interface {
	Publish(ctx context.Context, job domain.Job) error
}
