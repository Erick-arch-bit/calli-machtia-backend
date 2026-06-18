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

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

var (
	ErrNotFound = errors.New("record not found")
)

func (r *UserRepo) FindByEmail(email string) (*models.User, error) {
	query := `SELECT id, name, email, password_hash, role, avatar_url, bio, created_at, updated_at FROM users WHERE email = $1`

	user := &models.User{}
	err := r.db.QueryRow(context.Background(), query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash,
		&user.Role, &user.AvatarURL, &user.Bio, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return user, nil
}

func (r *UserRepo) FindByID(id string) (*models.User, error) {
	query := `SELECT id, name, email, password_hash, role, avatar_url, bio, created_at, updated_at FROM users WHERE id = $1`

	user := &models.User{}
	err := r.db.QueryRow(context.Background(), query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash,
		&user.Role, &user.AvatarURL, &user.Bio, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return user, nil
}

func (r *UserRepo) Create(user *models.User) error {
	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `INSERT INTO users (id, name, email, password_hash, role, avatar_url, bio, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.Exec(context.Background(), query,
		user.ID, user.Name, user.Email, user.PasswordHash,
		user.Role, user.AvatarURL, user.Bio, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepo) Update(user *models.User) error {
	user.UpdatedAt = time.Now()

	query := `UPDATE users SET name=$1, email=$2, password_hash=$3, role=$4, avatar_url=$5, bio=$6, updated_at=$7 WHERE id=$8`

	_, err := r.db.Exec(context.Background(), query,
		user.Name, user.Email, user.PasswordHash, user.Role,
		user.AvatarURL, user.Bio, user.UpdatedAt, user.ID,
	)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func (r *UserRepo) List(role string, page, limit int) ([]models.User, int64, error) {
	offset := (page - 1) * limit

	var total int64
	countQuery := `SELECT COUNT(*) FROM users`
	args := []interface{}{}

	if role != "" {
		countQuery += ` WHERE role = $1`
		args = append(args, role)
	}

	err := r.db.QueryRow(context.Background(), countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	dataQuery := `SELECT id, name, email, password_hash, role, avatar_url, bio, created_at, updated_at FROM users`
	queryArgs := []interface{}{}

	if role != "" {
		dataQuery += ` WHERE role = $1`
		queryArgs = append(queryArgs, role)
	}

	dataQuery += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", len(queryArgs)+1)
	queryArgs = append(queryArgs, limit)
	dataQuery += ` OFFSET $` + fmt.Sprintf("%d", len(queryArgs)+1)
	queryArgs = append(queryArgs, offset)

	rows, err := r.db.Query(context.Background(), dataQuery, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash,
			&u.Role, &u.AvatarURL, &u.Bio, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}

	if users == nil {
		users = []models.User{}
	}

	return users, total, nil
}
