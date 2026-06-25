package models

import "time"

type Course struct {
	ID              string    `json:"id" db:"id"`
	InstructorID    string    `json:"instructor_id" db:"instructor_id"`
	InstructorName  string    `json:"instructor_name,omitempty"`
	Title           string    `json:"title" db:"title" binding:"required"`
	Slug            string    `json:"slug" db:"slug"`
	Description     *string   `json:"description,omitempty" db:"description"`
	ImageURL        *string   `json:"image_url,omitempty" db:"image_url"`
	Price           float64   `json:"price" db:"price"`
	Category        *string   `json:"category,omitempty" db:"category"`
	Tags            []string  `json:"tags,omitempty" db:"tags"`
	Published       bool      `json:"published" db:"published"`
	EnrollmentCount int64     `json:"enrollment_count,omitempty"`
	SEOTitle        *string   `json:"seo_title,omitempty" db:"seo_title"`
	SEODescription  *string   `json:"seo_description,omitempty" db:"seo_description"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}
