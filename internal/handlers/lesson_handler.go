package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/calli-machtia/backend/internal/models"
	"github.com/calli-machtia/backend/internal/repository"
)

type LessonHandler struct {
	lessonRepo *repository.LessonRepo
}

func NewLessonHandler(lessonRepo *repository.LessonRepo) *LessonHandler {
	return &LessonHandler{lessonRepo: lessonRepo}
}

type createModuleRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Order       int    `json:"order"`
}

type updateModuleRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Order       int    `json:"order"`
}

type addLessonRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Content     string `json:"content"`
	VideoURL    string `json:"video_url"`
	Duration    int    `json:"duration"`
	Order       int    `json:"order"`
	Free        bool   `json:"free"`
}

type updateLessonRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Content     string `json:"content"`
	VideoURL    string `json:"video_url"`
	Duration    int    `json:"duration"`
	Order       int    `json:"order"`
	Free        bool   `json:"free"`
}

func (h *LessonHandler) GetModules(c *gin.Context) {
	courseID := c.Param("id")

	modules, err := h.lessonRepo.GetModules(courseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al cargar módulos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": modules})
}

func (h *LessonHandler) GetModule(c *gin.Context) {
	moduleID := c.Param("moduleId")

	module, err := h.lessonRepo.GetModule(moduleID)
	if err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "módulo no encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al cargar módulo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": module})
}

func (h *LessonHandler) CreateModule(c *gin.Context) {
	courseID := c.Param("id")

	var req createModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "el título es requerido"})
		return
	}

	module := &models.Module{
		CourseID:    courseID,
		Title:       req.Title,
		Description: req.Description,
		Order:       req.Order,
	}

	if err := h.lessonRepo.CreateModule(module); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al crear módulo"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": module})
}

func (h *LessonHandler) UpdateModule(c *gin.Context) {
	moduleID := c.Param("moduleId")

	module, err := h.lessonRepo.GetModule(moduleID)
	if err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "módulo no encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al cargar módulo"})
		return
	}

	var req updateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datos inválidos"})
		return
	}

	if req.Title != "" {
		module.Title = req.Title
	}
	if req.Description != "" {
		module.Description = req.Description
	}
	if req.Order != 0 {
		module.Order = req.Order
	}

	if err := h.lessonRepo.UpdateModule(module); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al actualizar módulo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": module})
}

func (h *LessonHandler) DeleteModule(c *gin.Context) {
	moduleID := c.Param("moduleId")

	if err := h.lessonRepo.DeleteModule(moduleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al eliminar módulo"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "módulo eliminado"}})
}

func (h *LessonHandler) AddLesson(c *gin.Context) {
	moduleID := c.Param("moduleId")

	var req addLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "el título es requerido"})
		return
	}

	lesson := &models.Lesson{
		Title:       req.Title,
		Description: req.Description,
		Content:     req.Content,
		VideoURL:    req.VideoURL,
		Duration:    req.Duration,
		Order:       req.Order,
		Free:        req.Free,
	}

	if err := h.lessonRepo.AddLesson(moduleID, lesson); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al agregar lección"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": lesson})
}

func (h *LessonHandler) UpdateLesson(c *gin.Context) {
	moduleID := c.Param("moduleId")
	lessonID := c.Param("lessonId")

	var req updateLessonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datos inválidos"})
		return
	}

	lesson := &models.Lesson{
		Title:       req.Title,
		Description: req.Description,
		Content:     req.Content,
		VideoURL:    req.VideoURL,
		Duration:    req.Duration,
		Order:       req.Order,
		Free:        req.Free,
	}

	if err := h.lessonRepo.UpdateLesson(moduleID, lessonID, lesson); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al actualizar lección"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": lesson})
}

func (h *LessonHandler) DeleteLesson(c *gin.Context) {
	moduleID := c.Param("moduleId")
	lessonID := c.Param("lessonId")

	if err := h.lessonRepo.DeleteLesson(moduleID, lessonID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al eliminar lección"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "lección eliminada"}})
}
