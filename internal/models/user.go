package models

import "time"

type User struct {
	ID           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name" binding:"required"`
	Email        string    `json:"email" db:"email" binding:"required,email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         string    `json:"role" db:"role"`
	AvatarURL    *string   `json:"avatar_url,omitempty" db:"avatar_url"`
	Bio          *string   `json:"bio,omitempty" db:"bio"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

const (
	RoleAlumno     = "alumno"
	RoleInstructor = "instructor"
	RoleAdmin      = "admin"
)
