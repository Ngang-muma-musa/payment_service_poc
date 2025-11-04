package handler

import (
	"errors"
	"fmt"
	"net/http"
	"paymentservice/internal/app/application"
	"paymentservice/internal/domain"

	"github.com/labstack/echo/v4"
)

type Payment interface {
	ProcessPayment(c echo.Context) error
}

type PaymentServiceHandler struct {
	paymentServiceApp application.PaymentServiceApp
}

func NewPaymentServiceHandler(
	paymentServiceApp application.PaymentServiceApp,
) Payment {
	return &PaymentServiceHandler{
		paymentServiceApp: paymentServiceApp,
	}
}

func (p *PaymentServiceHandler) ProcessPayment(c echo.Context) error {
	r := new(domain.Payment)
	if err := c.Bind(r); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": fmt.Sprintf("Invalid payload: %v", err),
		})
	}

	if r.UserID == "" || r.Amount <= 0 || r.Currency == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "userId, amount, and currency are required",
		})
	}
	payment, err := p.paymentServiceApp.CreateAndQueuePayment(
		c.Request().Context(),
		r.UserID,
		r.Amount,
		r.Currency,
	)

	if err != nil {
		if errors.Is(err, application.ErrRateLimitExceeded) {
			return c.JSON(http.StatusTooManyRequests, map[string]string{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Error occurred creating payment",
		})
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"status":  "success",
		"data":    payment,
		"message": "Payment queued successfully",
	})
}
