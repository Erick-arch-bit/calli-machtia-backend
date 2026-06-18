package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/calli-machtia/backend/internal/repository"
)

type AdminHandler struct {
	userRepo     *repository.UserRepo
	courseRepo   *repository.CourseRepo
	paymentRepo  *repository.PaymentRepo
	enrollRepo   *repository.EnrollmentRepo
}

func NewAdminHandler(
	userRepo *repository.UserRepo,
	courseRepo *repository.CourseRepo,
	paymentRepo *repository.PaymentRepo,
	enrollRepo *repository.EnrollmentRepo,
) *AdminHandler {
	return &AdminHandler{
		userRepo:    userRepo,
		courseRepo:  courseRepo,
		paymentRepo: paymentRepo,
		enrollRepo:  enrollRepo,
	}
}

type updateRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	role := c.Query("role")

	users, total, err := h.userRepo.List(role, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al cargar usuarios"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": users,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	userID := c.Param("id")

	var req updateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role es requerido"})
		return
	}

	if req.Role != "alumno" && req.Role != "instructor" && req.Role != "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rol inválido"})
		return
	}

	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "usuario no encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al buscar usuario"})
		return
	}

	user.Role = req.Role
	if err := h.userRepo.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al actualizar rol"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":   user.ID,
			"name": user.Name,
			"role": user.Role,
		},
	})
}

func (h *AdminHandler) ListCourses(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	filters := map[string]string{
		"published": c.Query("published"),
	}

	courses, total, err := h.courseRepo.FindAll(filters, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al cargar cursos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": courses,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

func (h *AdminHandler) Stats(c *gin.Context) {
	type statsResponse struct {
		TotalUsers       int64   `json:"total_users"`
		TotalCourses     int64   `json:"total_courses"`
		TotalRevenue     float64 `json:"total_revenue"`
		TotalEnrollments int64   `json:"total_enrollments"`
	}

	_, totalUsers, _ := h.userRepo.List("", 1, 1)
	_, totalCourses, _ := h.courseRepo.FindAll(nil, 1, 1)
	_, totalRevenue, _ := h.paymentRepo.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"data": statsResponse{
			TotalUsers:       totalUsers,
			TotalCourses:     totalCourses,
			TotalRevenue:     totalRevenue,
			TotalEnrollments: 0,
		},
	})
}

func (h *AdminHandler) ForceDeleteCourse(c *gin.Context) {
	id := c.Param("id")

	if err := h.courseRepo.ForceDelete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al eliminar curso"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "curso eliminado"}})
}
