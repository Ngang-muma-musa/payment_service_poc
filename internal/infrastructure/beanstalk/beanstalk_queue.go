package beanstalk

import (
	"context"
	"paymentservice/internal/domain"
	"time"

	"github.com/beanstalkd/go-beanstalk"
)

type BeanstalkQueue struct {
	tube *beanstalk.Tube
}

func NewBeanstalkQueue(conn *beanstalk.Conn, tubeName string) *BeanstalkQueue {
	return &BeanstalkQueue{
		tube: &beanstalk.Tube{Conn: conn, Name: tubeName},
	}
}

func (q *BeanstalkQueue) Publish(ctx context.Context, job domain.Job) error {
	_, err := q.tube.Put(job.Payload, 1, 0, 30*time.Second)
	return err
}
