package app

import (
	"paymentservice/internal/infrastructure/beanstalk"
	"paymentservice/internal/infrastructure/orm"
)

func Run() {
	paymentRepo := orm.NewPaymentServiceRepository()
	queue := beanstalk.NewBeanstalkQueue()
}
