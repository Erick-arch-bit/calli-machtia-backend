package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/calli-machtia/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CourseRepo struct {
	db *pgxpool.Pool
}

func NewCourseRepo(db *pgxpool.Pool) *CourseRepo {
	return &CourseRepo{db: db}
}

func (r *CourseRepo) FindAll(filters map[string]string, page, limit int) ([]models.Course, int64, error) {
	offset := (page - 1) * limit
	where := []string{}
	args := []interface{}{}
	argIdx := 1

	if filters == nil {
		filters = make(map[string]string)
	}

	if pub, ok := filters["published"]; ok && pub != "" {
		where = append(where, fmt.Sprintf("published = $%d", argIdx))
		args = append(args, pub == "true")
		argIdx++
	} else if _, ok := filters["published"]; !ok {
		where = append(where, fmt.Sprintf("published = $%d", argIdx))
		args = append(args, true)
		argIdx++
	}

	if cat, ok := filters["category"]; ok && cat != "" {
		where = append(where, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, cat)
		argIdx++
	}

	if search, ok := filters["search"]; ok && search != "" {
		where = append(where, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", argIdx, argIdx+1))
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern)
		argIdx += 2
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = " WHERE " + strings.Join(where, " AND ")
	}

	var total int64
	countQuery := "SELECT COUNT(*) FROM courses" + whereClause
	err := r.db.QueryRow(context.Background(), countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count courses: %w", err)
	}

	dataQuery := "SELECT id, instructor_id, title, slug, description, image_url, price, category, published, seo_title, seo_description, created_at, updated_at FROM courses" +
		whereClause + " ORDER BY created_at DESC LIMIT $" + fmt.Sprintf("%d", argIdx) + " OFFSET $" + fmt.Sprintf("%d", argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(context.Background(), dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list courses: %w", err)
	}
	defer rows.Close()

	var courses []models.Course
	for rows.Next() {
		var c models.Course
		err := rows.Scan(&c.ID, &c.InstructorID, &c.Title, &c.Slug,
			&c.Description, &c.ImageURL, &c.Price, &c.Category,
			&c.Published, &c.SEOTitle, &c.SEODescription, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, 0, fmt.Errorf("scan course: %w", err)
		}
		courses = append(courses, c)
	}

	if courses == nil {
		courses = []models.Course{}
	}

	return courses, total, nil
}

func (r *CourseRepo) FindBySlug(slug string) (*models.Course, error) {
	query := `SELECT id, instructor_id, title, slug, description, image_url, price, category, published, seo_title, seo_description, created_at, updated_at FROM courses WHERE slug = $1`

	course := &models.Course{}
	err := r.db.QueryRow(context.Background(), query, slug).Scan(
		&course.ID, &course.InstructorID, &course.Title, &course.Slug,
		&course.Description, &course.ImageURL, &course.Price, &course.Category,
		&course.Published, &course.SEOTitle, &course.SEODescription, &course.CreatedAt, &course.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find course by slug: %w", err)
	}
	return course, nil
}

func (r *CourseRepo) FindByID(id string) (*models.Course, error) {
	query := `SELECT id, instructor_id, title, slug, description, image_url, price, category, published, seo_title, seo_description, created_at, updated_at FROM courses WHERE id = $1`

	course := &models.Course{}
	err := r.db.QueryRow(context.Background(), query, id).Scan(
		&course.ID, &course.InstructorID, &course.Title, &course.Slug,
		&course.Description, &course.ImageURL, &course.Price, &course.Category,
		&course.Published, &course.SEOTitle, &course.SEODescription, &course.CreatedAt, &course.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("find course by id: %w", err)
	}
	return course, nil
}

func (r *CourseRepo) FindByInstructorID(instructorID string) ([]models.Course, error) {
	query := `SELECT id, instructor_id, title, slug, description, image_url, price, category, published, seo_title, seo_description, created_at, updated_at FROM courses WHERE instructor_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(context.Background(), query, instructorID)
	if err != nil {
		return nil, fmt.Errorf("find courses by instructor: %w", err)
	}
	defer rows.Close()

	var courses []models.Course
	for rows.Next() {
		var c models.Course
		err := rows.Scan(&c.ID, &c.InstructorID, &c.Title, &c.Slug,
			&c.Description, &c.ImageURL, &c.Price, &c.Category,
			&c.Published, &c.SEOTitle, &c.SEODescription, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan course: %w", err)
		}
		courses = append(courses, c)
	}

	if courses == nil {
		courses = []models.Course{}
	}

	return courses, nil
}

func (r *CourseRepo) Create(course *models.Course) error {
	course.ID = uuid.New().String()
	course.CreatedAt = time.Now()
	course.UpdatedAt = time.Now()

	query := `INSERT INTO courses (id, instructor_id, title, slug, description, image_url, price, category, published, seo_title, seo_description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.db.Exec(context.Background(), query,
		course.ID, course.InstructorID, course.Title, course.Slug,
		course.Description, course.ImageURL, course.Price, course.Category,
		course.Published, course.SEOTitle, course.SEODescription,
		course.CreatedAt, course.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create course: %w", err)
	}
	return nil
}

func (r *CourseRepo) Update(course *models.Course) error {
	course.UpdatedAt = time.Now()

	query := `UPDATE courses SET title=$1, slug=$2, description=$3, image_url=$4, price=$5, category=$6, published=$7, seo_title=$8, seo_description=$9, updated_at=$10 WHERE id=$11`

	_, err := r.db.Exec(context.Background(), query,
		course.Title, course.Slug, course.Description, course.ImageURL,
		course.Price, course.Category, course.Published, course.SEOTitle,
		course.SEODescription, course.UpdatedAt, course.ID,
	)
	if err != nil {
		return fmt.Errorf("update course: %w", err)
	}
	return nil
}

func (r *CourseRepo) Delete(id, instructorID string) error {
	query := `DELETE FROM courses WHERE id = $1 AND (instructor_id = $2 OR $2 = 'admin')`
	_, err := r.db.Exec(context.Background(), query, id, instructorID)
	if err != nil {
		return fmt.Errorf("delete course: %w", err)
	}
	return nil
}

func (r *CourseRepo) ForceDelete(id string) error {
	_, err := r.db.Exec(context.Background(), `DELETE FROM courses WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("force delete course: %w", err)
	}
	return nil
}
