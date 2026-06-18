package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PasswordResetRepo struct {
	db *pgxpool.Pool
}

func NewPasswordResetRepo(db *pgxpool.Pool) *PasswordResetRepo {
	return &PasswordResetRepo{db: db}
}

func (r *PasswordResetRepo) CreateToken(email string) (string, time.Time, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", time.Time{}, fmt.Errorf("generate token: %w", err)
	}
	token := hex.EncodeToString(bytes)
	expiresAt := time.Now().Add(1 * time.Hour)

	_, err := r.db.Exec(context.Background(),
		`INSERT INTO password_resets (email, token, expires_at) VALUES ($1, $2, $3)`,
		email, token, expiresAt,
	)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("save reset token: %w", err)
	}

	return token, expiresAt, nil
}

func (r *PasswordResetRepo) ValidateToken(token string) (string, error) {
	var email string
	var expiresAt time.Time
	var used bool

	err := r.db.QueryRow(context.Background(),
		`SELECT email, expires_at, used FROM password_resets WHERE token = $1`,
		token,
	).Scan(&email, &expiresAt, &used)
	if err != nil {
		return "", fmt.Errorf("token not found")
	}

	if used {
		return "", fmt.Errorf("token already used")
	}

	if time.Now().After(expiresAt) {
		return "", fmt.Errorf("token expired")
	}

	return email, nil
}

func (r *PasswordResetRepo) MarkTokenUsed(token string) error {
	_, err := r.db.Exec(context.Background(),
		`UPDATE password_resets SET used = true WHERE token = $1`,
		token,
	)
	return err
}

func (r *PasswordResetRepo) InvalidateExistingTokens(email string) error {
	_, err := r.db.Exec(context.Background(),
		`UPDATE password_resets SET used = true WHERE email = $1 AND used = false`,
		email,
	)
	return err
}
