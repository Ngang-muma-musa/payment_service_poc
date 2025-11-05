package orm

import (
	"context"
	"encoding/json"
	"errors"
	"paymentservice/internal/domain"
	"time"

	"github.com/go-redis/redis/v8"
)

var ErrPaymentNotFound = errors.New("payment not found")

type RedisPaymentRepository struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisPaymentRepository creates a new repository backed by Redis
func NewRedisPaymentRepository(client *redis.Client, ttl time.Duration) domain.Repository {
	return &RedisPaymentRepository{
		client: client,
		ttl:    ttl,
	}
}

// Create stores a payment in Redis
func (r *RedisPaymentRepository) Create(ctx context.Context, p *domain.Payment) error {
	p.Status = domain.StatusPending
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, p.ID, data, r.ttl).Err()
}

// UpdateStatus updates the status of a payment
func (r *RedisPaymentRepository) UpdateStatus(ctx context.Context, id string, status domain.PaymentStatus) error {
	p, err := r.FindByID(ctx, id)
	if err != nil {
		return err
	}

	p.Status = status
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, id, data, r.ttl).Err()
}

// FindByID retrieves a payment by ID
func (r *RedisPaymentRepository) FindByID(ctx context.Context, id string) (*domain.Payment, error) {
	data, err := r.client.Get(ctx, id).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}

	var p domain.Payment
	if err := json.Unmarshal([]byte(data), &p); err != nil {
		return nil, err
	}

	return &p, nil
}
