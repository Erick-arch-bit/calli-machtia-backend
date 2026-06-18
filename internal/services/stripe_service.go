package services

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/calli-machtia/backend/internal/config"
	"github.com/calli-machtia/backend/internal/models"
	"github.com/calli-machtia/backend/internal/repository"
	"github.com/stripe/stripe-go/v81"
	stripeclient "github.com/stripe/stripe-go/v81/client"
	"github.com/stripe/stripe-go/v81/paymentintent"
	"github.com/stripe/stripe-go/v81/webhook"
)

type StripeService struct {
	cfg         *config.Config
	paymentRepo *repository.PaymentRepo
	enrollRepo  *repository.EnrollmentRepo
	client      *stripeclient.API
}

func NewStripeService(cfg *config.Config, paymentRepo *repository.PaymentRepo, enrollRepo *repository.EnrollmentRepo) *StripeService {
	sc := &stripeclient.API{}
	sc.Init(cfg.STRIPE_SECRET_KEY, nil)
	return &StripeService{
		cfg:         cfg,
		paymentRepo: paymentRepo,
		enrollRepo:  enrollRepo,
		client:      sc,
	}
}

func (s *StripeService) CreatePaymentIntent(amount float64, currency, courseID, userID string) (*stripe.PaymentIntent, error) {
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(int64(amount * 100)),
		Currency: stripe.String(currency),
		Metadata: map[string]string{
			"course_id": courseID,
			"user_id":   userID,
		},
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled: stripe.Bool(true),
		},
	}

	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("create payment intent: %w", err)
	}

	payment := &models.Payment{
		UserID:          userID,
		CourseID:        courseID,
		Amount:          amount,
		Currency:        currency,
		StripePaymentID: pi.ID,
		Status:          models.PaymentPending,
	}

	if err := s.paymentRepo.Create(payment); err != nil {
		return nil, fmt.Errorf("save payment: %w", err)
	}

	return pi, nil
}

func (s *StripeService) ConstructEvent(payload []byte, sigHeader string) (stripe.Event, error) {
	return webhook.ConstructEvent(payload, sigHeader, s.cfg.STRIPE_WEBHOOK_SECRET)
}

func (s *StripeService) HandleWebhook(payload []byte, sigHeader string) error {
	event, err := s.ConstructEvent(payload, sigHeader)
	if err != nil {
		return fmt.Errorf("webhook signature verification failed: %w", err)
	}

	switch event.Type {
	case "payment_intent.succeeded":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			return fmt.Errorf("unmarshal payment intent: %w", err)
		}
		return s.handlePaymentSucceeded(&pi)

	case "payment_intent.payment_failed":
		var pi stripe.PaymentIntent
		if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
			return fmt.Errorf("unmarshal payment intent: %w", err)
		}
		return s.handlePaymentFailed(&pi)

	default:
		log.Printf("Unhandled webhook event: %s", event.Type)
	}

	return nil
}

func (s *StripeService) handlePaymentSucceeded(pi *stripe.PaymentIntent) error {
	payment, err := s.paymentRepo.FindByStripePaymentID(pi.ID)
	if err != nil {
		return fmt.Errorf("find payment by stripe id: %w", err)
	}

	if err := s.paymentRepo.UpdateStatus(payment.ID, models.PaymentCompleted); err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}

	enrollment := &models.Enrollment{
		UserID:   pi.Metadata["user_id"],
		CourseID: pi.Metadata["course_id"],
		Status:   models.EnrollmentActive,
		Progress: 0,
	}

	if err := s.enrollRepo.Create(enrollment); err != nil {
		return fmt.Errorf("create enrollment from payment: %w", err)
	}

	return nil
}

func (s *StripeService) handlePaymentFailed(pi *stripe.PaymentIntent) error {
	payment, err := s.paymentRepo.FindByStripePaymentID(pi.ID)
	if err != nil {
		return fmt.Errorf("find payment by stripe id: %w", err)
	}

	return s.paymentRepo.UpdateStatus(payment.ID, models.PaymentFailed)
}
