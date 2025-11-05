package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"paymentservice/internal/app/application"
	queue "paymentservice/internal/infrastructure/beanstalk"
	"paymentservice/internal/infrastructure/orm"
	"paymentservice/internal/infrastructure/redis"
	"paymentservice/internal/presentation/restapi/handler"
	"paymentservice/internal/presentation/router"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
)

func Run(
	beanstalkAddr string,
	beanstalkTubeName string,
	redisDsn string,
	rateLimit int64,
	rateLimitWindow int64,
	appPort int64,
) {

	// Initialize Beanstalk queue
	conn, err := queue.BuildBeanstalkQueue(beanstalkAddr)
	if err != nil {
		log.Fatalf("failed to connect to beanstalk: %v", err)
	}
	defer conn.Close()

	beanstalkQueue := queue.NewBeanstalkQueue(conn, beanstalkTubeName)

	// Initialize Redis and rate limiter
	cache, err := redis.BuildRedisCache(redisDsn)
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	defer cache.Close()

	window := time.Duration(rateLimitWindow) * time.Minute
	ratelimiter := redis.NewRedisRateLimiter(cache, rateLimit, window)

	// Initialize repository
	paymentRepo := orm.NewRedisPaymentRepository(cache, 0)

	// Root context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Application layer
	paymentServiceApp := application.NewPaymentService(
		paymentRepo,
		beanstalkQueue,
		ratelimiter,
	)

	// Presentation layer
	paymentServiceHandler := handler.NewPaymentServiceHandler(paymentServiceApp)
	apiRouter := router.NewRouter(paymentServiceHandler)

	e := echo.New()
	apiRouter.Register(e)
	echoServer := router.NewServer(e, appPort)

	// Start server
	echoServer.Start()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctxShutdown, cancelShutdown := context.WithTimeout(ctx, 10*time.Second)
	defer cancelShutdown()

	echoServer.Shutdown(ctxShutdown)
}
