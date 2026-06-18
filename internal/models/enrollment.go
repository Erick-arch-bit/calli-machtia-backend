package models

import "time"

type Enrollment struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	CourseID  string    `json:"course_id" db:"course_id"`
	Status    string    `json:"status" db:"status"`
	Progress  float64   `json:"progress" db:"progress"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

const (
	EnrollmentActive    = "active"
	EnrollmentCompleted = "completed"
	EnrollmentCancelled = "cancelled"
)
