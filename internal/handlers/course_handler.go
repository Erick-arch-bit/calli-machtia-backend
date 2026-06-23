package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/calli-machtia/backend/internal/middleware"
	"github.com/calli-machtia/backend/internal/models"
	"github.com/calli-machtia/backend/internal/repository"
)

type CourseHandler struct {
	courseRepo *repository.CourseRepo
	lessonRepo *repository.LessonRepo
}

func NewCourseHandler(courseRepo *repository.CourseRepo, lessonRepo *repository.LessonRepo) *CourseHandler {
	return &CourseHandler{courseRepo: courseRepo, lessonRepo: lessonRepo}
}

type createCourseRequest struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description"`
	ImageURL    string   `json:"image_url"`
	Price       float64  `json:"price"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
}

type updateCourseRequest struct {
	Title          *string   `json:"title"`
	Description    *string   `json:"description"`
	ImageURL       *string   `json:"image_url"`
	Price          *float64  `json:"price"`
	Category       *string   `json:"category"`
	Tags           *[]string `json:"tags"`
	Published      *bool     `json:"published"`
	SEOTitle       *string   `json:"seo_title"`
	SEODescription *string   `json:"seo_description"`
}

func slugify(text string) string {
	slug := strings.ToLower(text)
	slug = strings.NewReplacer(
		"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u",
		"ñ", "n", "ü", "u",
	).Replace(slug)
	slug = strings.TrimSpace(slug)
	result := make([]byte, 0, len(slug))
	lastWasDash := false
	for _, c := range []byte(slug) {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			result = append(result, c)
			lastWasDash = false
		} else if c == ' ' || c == '-' || c == '_' {
			if !lastWasDash {
				result = append(result, '-')
				lastWasDash = true
			}
		}
	}
	return string(result)
}

func (h *CourseHandler) List(c *gin.Context) {
	filters := make(map[string]string)
	if cat := c.Query("category"); cat != "" {
		filters["category"] = cat
	}
	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}
	if pub := c.Query("published"); pub != "" {
		filters["published"] = pub
	}

	page := 1
	limit := 12
	if p, err := parseInt(c.Query("page")); err == nil && p > 0 {
		page = p
	}
	if l, err := parseInt(c.Query("limit")); err == nil && l > 0 && l <= 50 {
		limit = l
	}

	courses, total, err := h.courseRepo.FindAll(filters, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al cargar cursos"})
		return
	}

	pages := int(total) / limit
	if int(total)%limit > 0 {
		pages++
	}

	c.JSON(http.StatusOK, gin.H{
		"data": courses,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
			"pages": pages,
		},
	})
}

func parseInt(s string) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("not a number")
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

func (h *CourseHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")
	course, err := h.courseRepo.FindBySlug(slug)
	if err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "curso no encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al buscar curso"})
		return
	}

	modules, _ := h.lessonRepo.GetModules(course.ID)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"course":  course,
			"modules": modules,
		},
	})
}

func (h *CourseHandler) Create(c *gin.Context) {
	var req createCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "el título es requerido"})
		return
	}

	slug := slugify(req.Title)
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "el título no genera un slug válido"})
		return
	}

	description := &req.Description
	if req.Description == "" {
		description = nil
	}
	imageURL := &req.ImageURL
	if req.ImageURL == "" {
		imageURL = nil
	}
	category := &req.Category
	if req.Category == "" {
		category = nil
	}
	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}

	course := &models.Course{
		InstructorID: middleware.GetUserID(c),
		Title:        req.Title,
		Slug:         slug,
		Description:  description,
		ImageURL:     imageURL,
		Price:        req.Price,
		Category:     category,
		Tags:         tags,
		Published:    false,
	}

	if err := h.courseRepo.Create(course); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al crear curso"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": course})
}

func (h *CourseHandler) Update(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	role := middleware.GetRole(c)

	course, err := h.courseRepo.FindByID(id)
	if err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "curso no encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al buscar curso"})
		return
	}

	if course.InstructorID != userID && role != models.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "no tienes permiso para modificar este curso"})
		return
	}

	var req updateCourseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datos inválidos"})
		return
	}

	if req.Title != nil {
		course.Title = *req.Title
		course.Slug = slugify(*req.Title)
	}
	if req.Description != nil {
		if *req.Description == "" {
			course.Description = nil
		} else {
			course.Description = req.Description
		}
	}
	if req.ImageURL != nil {
		if *req.ImageURL == "" {
			course.ImageURL = nil
		} else {
			course.ImageURL = req.ImageURL
		}
	}
	if req.Price != nil {
		course.Price = *req.Price
	}
	if req.Category != nil {
		if *req.Category == "" {
			course.Category = nil
		} else {
			course.Category = req.Category
		}
	}
	if req.Tags != nil {
		course.Tags = *req.Tags
	}
	if req.Published != nil {
		course.Published = *req.Published
	}
	if req.SEOTitle != nil {
		course.SEOTitle = req.SEOTitle
	}
	if req.SEODescription != nil {
		course.SEODescription = req.SEODescription
	}

	if err := h.courseRepo.Update(course); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al actualizar curso"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": course})
}

func (h *CourseHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)

	if err := h.courseRepo.Delete(id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al eliminar curso"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "curso eliminado"}})
}

func (h *CourseHandler) ListCategories(c *gin.Context) {
	categories, err := h.courseRepo.FindCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al cargar categorías"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": categories})
}

func (h *CourseHandler) MyCourses(c *gin.Context) {
	userID := middleware.GetUserID(c)

	courses, err := h.courseRepo.FindByInstructorID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al cargar cursos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": courses})
}
