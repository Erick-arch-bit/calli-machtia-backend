package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/calli-machtia/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepo struct {
	db *pgxpool.Pool
}

func NewPaymentRepo(db *pgxpool.Pool) *PaymentRepo {
	return &PaymentRepo{db: db}
}

func (r *PaymentRepo) Create(payment *models.Payment) error {
	payment.ID = uuid.New().String()
	payment.CreatedAt = time.Now()

	query := `INSERT INTO payments (id, user_id, course_id, amount, currency, stripe_payment_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.Exec(context.Background(), query,
		payment.ID, payment.UserID, payment.CourseID, payment.Amount,
		payment.Currency, payment.StripePaymentID, payment.Status, payment.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create payment: %w", err)
	}
	return nil
}

func (r *PaymentRepo) FindByID(id string) (*models.Payment, error) {
	query := `SELECT id, user_id, course_id, amount, currency, stripe_payment_id, status, created_at FROM payments WHERE id = $1`

	p := &models.Payment{}
	err := r.db.QueryRow(context.Background(), query, id).Scan(
		&p.ID, &p.UserID, &p.CourseID, &p.Amount,
		&p.Currency, &p.StripePaymentID, &p.Status, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find payment by id: %w", err)
	}
	return p, nil
}

func (r *PaymentRepo) FindByUser(userID string) ([]models.Payment, error) {
	query := `SELECT id, user_id, course_id, amount, currency, stripe_payment_id, status, created_at FROM payments WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("find payments by user: %w", err)
	}
	defer rows.Close()

	var payments []models.Payment
	for rows.Next() {
		var p models.Payment
		err := rows.Scan(&p.ID, &p.UserID, &p.CourseID, &p.Amount,
			&p.Currency, &p.StripePaymentID, &p.Status, &p.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan payment: %w", err)
		}
		payments = append(payments, p)
	}

	if payments == nil {
		payments = []models.Payment{}
	}

	return payments, nil
}

func (r *PaymentRepo) FindAll(page, limit int) ([]models.Payment, int64, error) {
	offset := (page - 1) * limit

	var total int64
	err := r.db.QueryRow(context.Background(), `SELECT COUNT(*) FROM payments`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count payments: %w", err)
	}

	query := `SELECT id, user_id, course_id, amount, currency, stripe_payment_id, status, created_at FROM payments ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(context.Background(), query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list payments: %w", err)
	}
	defer rows.Close()

	var payments []models.Payment
	for rows.Next() {
		var p models.Payment
		err := rows.Scan(&p.ID, &p.UserID, &p.CourseID, &p.Amount,
			&p.Currency, &p.StripePaymentID, &p.Status, &p.CreatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("scan payment: %w", err)
		}
		payments = append(payments, p)
	}

	if payments == nil {
		payments = []models.Payment{}
	}

	return payments, total, nil
}

func (r *PaymentRepo) GetTotalRevenue() (float64, error) {
	var total float64
	err := r.db.QueryRow(context.Background(), `SELECT COALESCE(SUM(amount), 0) FROM payments WHERE status = 'completed'`).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("get total revenue: %w", err)
	}
	return total, nil
}

func (r *PaymentRepo) FindByStripePaymentID(stripeID string) (*models.Payment, error) {
	query := `SELECT id, user_id, course_id, amount, currency, stripe_payment_id, status, created_at FROM payments WHERE stripe_payment_id = $1`

	p := &models.Payment{}
	err := r.db.QueryRow(context.Background(), query, stripeID).Scan(
		&p.ID, &p.UserID, &p.CourseID, &p.Amount,
		&p.Currency, &p.StripePaymentID, &p.Status, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find payment by stripe id: %w", err)
	}
	return p, nil
}

func (r *PaymentRepo) GetStats() (int64, float64, error) {
	var count int64
	var total float64
	err := r.db.QueryRow(context.Background(), `SELECT COUNT(*), COALESCE(SUM(amount), 0) FROM payments WHERE status = 'completed'`).Scan(&count, &total)
	if err != nil {
		return 0, 0, fmt.Errorf("get payment stats: %w", err)
	}
	return count, total, nil
}

func (r *PaymentRepo) UpdateStatus(id, status string) error {
	_, err := r.db.Exec(context.Background(), `UPDATE payments SET status=$1 WHERE id=$2`, status, id)
	if err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}
	return nil
}
