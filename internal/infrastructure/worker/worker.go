package worker

import (
	"context"
	"encoding/json"
	"log"
	"paymentservice/internal/domain"
	"time"

	"github.com/beanstalkd/go-beanstalk"
)

type Worker struct {
	ctx      context.Context
	conn     *beanstalk.Conn
	repo     domain.Repository
	tubeName string
}

func NewWorker(
	ctx context.Context,
	conn *beanstalk.Conn,
	repo domain.Repository,
	tubeName string,
) *Worker {
	return &Worker{
		ctx:      ctx,
		conn:     conn,
		repo:     repo,
		tubeName: tubeName,
	}
}

func (w *Worker) Start(ctx context.Context) {
	log.Println("Worker started...")

	tube := beanstalk.NewTubeSet(w.conn, w.tubeName)

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker shutting down gracefully...")
			return

		default:
			id, body, err := tube.Reserve(24 * time.Hour)
			if err != nil {
				log.Printf("Error reserving job: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			var payment domain.Payment
			if err := json.Unmarshal(body, &payment); err != nil {
				log.Printf("failed to deserialize job: %v", err)
				w.conn.Bury(id, 0)
				continue
			}

			log.Printf("Processing payment ID: %s for user %s amount: %.2f %s",
				payment.ID, payment.UserID, payment.Amount, payment.Currency)

			// Simulate payment processing
			time.Sleep(3 * time.Second)
			payment.Status = "PROCESSED"
			w.repo.UpdateStatus(ctx, payment.ID, payment.Status)

			w.conn.Delete(id)
			log.Printf("Payment %s processed successfully", payment.ID)
		}
	}
}
