package models

import "time"

type Module struct {
	ID          string    `json:"id" bson:"_id"`
	CourseID    string    `json:"course_id" bson:"course_id"`
	Title       string    `json:"title" bson:"title"`
	Description string    `json:"description" bson:"description"`
	Order       int       `json:"order" bson:"order"`
	Lessons     []Lesson  `json:"lessons" bson:"lessons"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
}

type Lesson struct {
	ID          string `json:"id" bson:"id"`
	Title       string `json:"title" bson:"title"`
	Description string `json:"description" bson:"description"`
	Content     string `json:"content" bson:"content"`
	VideoURL    string `json:"video_url,omitempty" bson:"video_url,omitempty"`
	Duration    int    `json:"duration" bson:"duration"`
	Order       int    `json:"order" bson:"order"`
	Free        bool   `json:"free" bson:"free"`
}
