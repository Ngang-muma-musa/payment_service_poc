package application

import (
	"context"
	"encoding/json"
	"errors"
	"paymentservice/internal/domain"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

var ErrRateLimitExceeded = errors.New("rate limit exceeded")

type PaymentServiceApp interface {
	CreateAndQueuePayment(
		ctx context.Context,
		userID string,
		amount float64,
		currency string,
	) (*domain.Payment, error)

	GetPaymentByID(
		ctx context.Context,
		paymentID string,
	) (*domain.Payment, error)
}

type PaymentService struct {
	repo        domain.Repository
	queue       QueuePort
	rateLimiter RateLimiterPort
}

func NewPaymentService(
	repo domain.Repository,
	queue QueuePort,
	rateLimiter RateLimiterPort,
) *PaymentService {
	return &PaymentService{
		repo:        repo,
		queue:       queue,
		rateLimiter: rateLimiter,
	}
}

// CreateAndQueuePayment handles:
// - rate limit check
// - create payment record
// - enqueue payment job
func (s *PaymentService) CreateAndQueuePayment(
	ctx context.Context,
	userID string,
	amount float64,
	currency string,
) (*domain.Payment, error) {
	allowed, err := s.rateLimiter.Allow(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrRateLimitExceeded
	}

	payment := &domain.Payment{
		ID:        gofakeit.UUID(),
		UserID:    userID,
		Amount:    amount,
		Currency:  currency,
		Status:    domain.StatusPending,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, payment); err != nil {
		return nil, err
	}

	p, err := json.Marshal(payment)
	if err != nil {
		return nil, err
	}

	job := domain.Job{
		ID:      payment.ID,
		Payload: p,
	}

	if err := s.queue.Publish(ctx, job); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateStatus(ctx, payment.ID, "QUEUED"); err != nil {
		return nil, err
	}

	return payment, nil
}

func (s *PaymentService) GetPaymentByID(
	ctx context.Context,
	paymentID string,
) (*domain.Payment, error) {
	return s.repo.FindByID(ctx, paymentID)
}
