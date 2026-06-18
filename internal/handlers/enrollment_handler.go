package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/calli-machtia/backend/internal/middleware"
	"github.com/calli-machtia/backend/internal/models"
	"github.com/calli-machtia/backend/internal/repository"
)

type EnrollmentHandler struct {
	enrollRepo *repository.EnrollmentRepo
	courseRepo *repository.CourseRepo
}

func NewEnrollmentHandler(enrollRepo *repository.EnrollmentRepo, courseRepo *repository.CourseRepo) *EnrollmentHandler {
	return &EnrollmentHandler{enrollRepo: enrollRepo, courseRepo: courseRepo}
}

type enrollRequest struct {
	CourseID string `json:"course_id" binding:"required"`
}

type progressRequest struct {
	Progress float64 `json:"progress" binding:"min=0,max=100"`
}

func (h *EnrollmentHandler) Enroll(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req enrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "course_id es requerido"})
		return
	}

	course, err := h.courseRepo.FindByID(req.CourseID)
	if err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "curso no encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al verificar curso"})
		return
	}

	if !course.Published {
		c.JSON(http.StatusBadRequest, gin.H{"error": "el curso no está disponible"})
		return
	}

	existing, err := h.enrollRepo.FindByUserAndCourse(userID, req.CourseID)
	if err != nil && err != repository.ErrNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al verificar inscripción"})
		return
	}
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "ya estás inscrito a este curso"})
		return
	}

	enrollment := &models.Enrollment{
		UserID:   userID,
		CourseID: req.CourseID,
		Status:   models.EnrollmentActive,
		Progress: 0,
	}

	if err := h.enrollRepo.Create(enrollment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al inscribirte"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": enrollment})
}

func (h *EnrollmentHandler) MyEnrollments(c *gin.Context) {
	userID := middleware.GetUserID(c)

	enrollments, err := h.enrollRepo.FindByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al cargar inscripciones"})
		return
	}

	type enrollmentWithCourse struct {
		models.Enrollment
		Course *models.Course `json:"course"`
	}

	result := make([]enrollmentWithCourse, 0, len(enrollments))
	for _, e := range enrollments {
		course, err := h.courseRepo.FindByID(e.CourseID)
		if err == nil && course != nil {
			result = append(result, enrollmentWithCourse{Enrollment: e, Course: course})
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
}

func (h *EnrollmentHandler) CourseEnrollments(c *gin.Context) {
	courseID := c.Param("courseId")
	userID := middleware.GetUserID(c)
	role := middleware.GetRole(c)

	if role != models.RoleAdmin {
		course, err := h.courseRepo.FindByID(courseID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "curso no encontrado"})
			return
		}
		if course.InstructorID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "no tienes permiso"})
			return
		}
	}

	enrollments, err := h.enrollRepo.FindByCourse(courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al cargar inscripciones"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": enrollments})
}

func (h *EnrollmentHandler) UpdateProgress(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)

	var req progressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "progreso debe ser entre 0 y 100"})
		return
	}

	e, err := h.enrollRepo.FindByID(id)
	if err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "inscripción no encontrada"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al buscar inscripción"})
		return
	}

	if e.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "no tienes permiso"})
		return
	}

	e.Progress = req.Progress
	if req.Progress >= 100 {
		e.Status = models.EnrollmentCompleted
	}

	if err := h.enrollRepo.Update(e); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al actualizar progreso"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": e})
}

func (h *EnrollmentHandler) Unenroll(c *gin.Context) {
	id := c.Param("id")

	if err := h.enrollRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al cancelar inscripción"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "inscripción cancelada"}})
}
