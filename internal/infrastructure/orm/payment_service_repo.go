package orm

import (
	"context"
	"errors"
	"paymentservice/internal/domain"
	"sync"
)

// PaymentServiceRepository implements domain.PaymentRepository
type PaymentServiceRepository struct {
	mu       sync.Mutex
	payments map[string]*domain.Payment
}

func NewPaymentServiceRepository() domain.Repository {
	return &PaymentServiceRepository{
		payments: make(map[string]*domain.Payment),
	}
}

func (r *PaymentServiceRepository) Create(
	ctx context.Context,
	p *domain.Payment,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p.Status = "PENDING"
	r.payments[p.ID] = p
	return nil
}

func (r *PaymentServiceRepository) UpdateStatus(
	ctx context.Context,
	id string,
	status domain.PaymentStatus,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.payments[id]
	if !ok {
		return errors.New("payment not found")
	}
	p.Status = status
	return nil
}

func (r *PaymentServiceRepository) FindByID(
	ctx context.Context,
	id string,
) (*domain.Payment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.payments[id]
	if !ok {
		return nil, errors.New("payment not found")
	}
	return p, nil
}
