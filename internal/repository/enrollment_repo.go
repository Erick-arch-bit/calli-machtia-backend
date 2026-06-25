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

type EnrollmentRepo struct {
	db *pgxpool.Pool
}

func NewEnrollmentRepo(db *pgxpool.Pool) *EnrollmentRepo {
	return &EnrollmentRepo{db: db}
}

func (r *EnrollmentRepo) FindByID(id string) (*models.Enrollment, error) {
	query := `SELECT id, user_id, course_id, status, progress, created_at, updated_at FROM enrollments WHERE id = $1`

	e := &models.Enrollment{}
	err := r.db.QueryRow(context.Background(), query, id).Scan(
		&e.ID, &e.UserID, &e.CourseID, &e.Status, &e.Progress, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find enrollment by id: %w", err)
	}
	return e, nil
}

func (r *EnrollmentRepo) Create(enrollment *models.Enrollment) error {
	enrollment.ID = uuid.New().String()
	enrollment.CreatedAt = time.Now()
	enrollment.UpdatedAt = time.Now()

	query := `INSERT INTO enrollments (id, user_id, course_id, status, progress, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.Exec(context.Background(), query,
		enrollment.ID, enrollment.UserID, enrollment.CourseID,
		enrollment.Status, enrollment.Progress, enrollment.CreatedAt, enrollment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create enrollment: %w", err)
	}
	return nil
}

func (r *EnrollmentRepo) FindByUserAndCourse(userID, courseID string) (*models.Enrollment, error) {
	query := `SELECT id, user_id, course_id, status, progress, created_at, updated_at FROM enrollments WHERE user_id = $1 AND course_id = $2`

	e := &models.Enrollment{}
	err := r.db.QueryRow(context.Background(), query, userID, courseID).Scan(
		&e.ID, &e.UserID, &e.CourseID, &e.Status, &e.Progress, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find enrollment: %w", err)
	}
	return e, nil
}

type EnrollmentWithCourse struct {
	models.Enrollment
	Course *models.Course `json:"course"`
}

func (r *EnrollmentRepo) FindByUser(userID string) ([]models.Enrollment, error) {
	query := `SELECT id, user_id, course_id, status, progress, created_at, updated_at FROM enrollments WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("find enrollments by user: %w", err)
	}
	defer rows.Close()

	var enrollments []models.Enrollment
	for rows.Next() {
		var e models.Enrollment
		err := rows.Scan(&e.ID, &e.UserID, &e.CourseID, &e.Status, &e.Progress, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan enrollment: %w", err)
		}
		enrollments = append(enrollments, e)
	}

	if enrollments == nil {
		enrollments = []models.Enrollment{}
	}

	return enrollments, nil
}

func (r *EnrollmentRepo) FindByUserWithCourses(userID string) ([]EnrollmentWithCourse, error) {
	query := `SELECT e.id, e.user_id, e.course_id, e.status, e.progress, e.created_at, e.updated_at,
		c.id, c.instructor_id, COALESCE(uc.name, ''), c.title, c.slug,
		c.description, c.image_url, c.price, c.category, c.tags, c.published,
		c.seo_title, c.seo_description, c.created_at, c.updated_at,
		COALESCE(ec.cnt, 0)
		FROM enrollments e
		INNER JOIN courses c ON c.id = e.course_id
		LEFT JOIN users uc ON uc.id = c.instructor_id
		LEFT JOIN (SELECT course_id, COUNT(*) as cnt FROM enrollments GROUP BY course_id) ec ON ec.course_id = c.id
		WHERE e.user_id = $1
		ORDER BY e.created_at DESC`

	rows, err := r.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("find enrollments with courses: %w", err)
	}
	defer rows.Close()

	var result []EnrollmentWithCourse
	for rows.Next() {
		var ew EnrollmentWithCourse
		ew.Course = &models.Course{}
		err := rows.Scan(
			&ew.ID, &ew.UserID, &ew.CourseID, &ew.Status, &ew.Progress, &ew.CreatedAt, &ew.UpdatedAt,
			&ew.Course.ID, &ew.Course.InstructorID, &ew.Course.InstructorName, &ew.Course.Title, &ew.Course.Slug,
			&ew.Course.Description, &ew.Course.ImageURL, &ew.Course.Price, &ew.Course.Category, &ew.Course.Tags,
			&ew.Course.Published, &ew.Course.SEOTitle, &ew.Course.SEODescription, &ew.Course.CreatedAt, &ew.Course.UpdatedAt,
			&ew.Course.EnrollmentCount,
		)
		if err != nil {
			return nil, fmt.Errorf("scan enrollment with course: %w", err)
		}
		result = append(result, ew)
	}

	if result == nil {
		result = []EnrollmentWithCourse{}
	}

	return result, nil
}

func (r *EnrollmentRepo) FindByCourse(courseID string) ([]models.Enrollment, error) {
	query := `SELECT id, user_id, course_id, status, progress, created_at, updated_at FROM enrollments WHERE course_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(context.Background(), query, courseID)
	if err != nil {
		return nil, fmt.Errorf("find enrollments by course: %w", err)
	}
	defer rows.Close()

	var enrollments []models.Enrollment
	for rows.Next() {
		var e models.Enrollment
		err := rows.Scan(&e.ID, &e.UserID, &e.CourseID, &e.Status, &e.Progress, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan enrollment: %w", err)
		}
		enrollments = append(enrollments, e)
	}

	if enrollments == nil {
		enrollments = []models.Enrollment{}
	}

	return enrollments, nil
}

func (r *EnrollmentRepo) Update(enrollment *models.Enrollment) error {
	enrollment.UpdatedAt = time.Now()

	query := `UPDATE enrollments SET status=$1, progress=$2, updated_at=$3 WHERE id=$4`

	_, err := r.db.Exec(context.Background(), query,
		enrollment.Status, enrollment.Progress, enrollment.UpdatedAt, enrollment.ID,
	)
	if err != nil {
		return fmt.Errorf("update enrollment: %w", err)
	}
	return nil
}

func (r *EnrollmentRepo) CountByCourse(courseID string) (int64, error) {
	query := `SELECT COUNT(*) FROM enrollments WHERE course_id = $1`
	var count int64
	err := r.db.QueryRow(context.Background(), query, courseID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count enrollments: %w", err)
	}
	return count, nil
}

func (r *EnrollmentRepo) Delete(id string) error {
	_, err := r.db.Exec(context.Background(), `DELETE FROM enrollments WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete enrollment: %w", err)
	}
	return nil
}
