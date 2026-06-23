package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/calli-machtia/backend/internal/middleware"
	"github.com/calli-machtia/backend/internal/models"
	"github.com/calli-machtia/backend/internal/repository"
	"github.com/calli-machtia/backend/internal/services"
)

type AuthHandler struct {
	authSvc        *services.AuthService
	userRepo       *repository.UserRepo
	resetRepo      *repository.PasswordResetRepo
}

func NewAuthHandler(authSvc *services.AuthService, userRepo *repository.UserRepo, resetRepo *repository.PasswordResetRepo) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, userRepo: userRepo, resetRepo: resetRepo}
}

type registerRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type updateProfileRequest struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Bio       string `json:"bio"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Role == "" {
		req.Role = models.RoleAlumno
	}
	if req.Role != models.RoleAlumno && req.Role != models.RoleInstructor && req.Role != models.RoleAdmin {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rol inválido: debe ser alumno o instructor"})
		return
	}

	existing, err := h.userRepo.FindByEmail(req.Email)
	if err != nil && err != repository.ErrNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al verificar email"})
		return
	}
	if existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "este email ya está registrado"})
		return
	}

	hash, err := h.authSvc.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al procesar contraseña"})
		return
	}

	user := &models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hash,
		Role:         req.Role,
	}

	if err := h.userRepo.Create(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al crear usuario"})
		return
	}

	accessToken, err := h.authSvc.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al generar token"})
		return
	}

	refreshToken, err := h.authSvc.GenerateRefreshToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al generar refresh token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": gin.H{
			"user": gin.H{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
				"role":  user.Role,
			},
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
		},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email y contraseña son requeridos"})
		return
	}

	user, err := h.userRepo.FindByEmail(req.Email)
	if err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "email o contraseña incorrectos"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al buscar usuario"})
		return
	}

	if !h.authSvc.CheckPassword(user.PasswordHash, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "email o contraseña incorrectos"})
		return
	}

	accessToken, err := h.authSvc.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al generar token"})
		return
	}

	refreshToken, err := h.authSvc.GenerateRefreshToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al generar refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"user": gin.H{
				"id":         user.ID,
				"name":       user.Name,
				"email":      user.Email,
				"role":       user.Role,
				"avatar_url": user.AvatarURL,
				"bio":        user.Bio,
			},
			"accessToken":  accessToken,
			"refreshToken": refreshToken,
		},
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)

	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		if err == repository.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "usuario no encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al buscar usuario"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"role":       user.Role,
			"avatar_url": user.AvatarURL,
			"bio":        user.Bio,
			"created_at": user.CreatedAt,
		},
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token es requerido"})
		return
	}

	claims, err := h.authSvc.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token inválido o expirado"})
		return
	}

	user, err := h.userRepo.FindByID(claims.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no encontrado"})
		return
	}

	accessToken, err := h.authSvc.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al generar token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"accessToken": accessToken,
		},
	})
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "datos inválidos"})
		return
	}

	user, err := h.userRepo.FindByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "usuario no encontrado"})
		return
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.AvatarURL != "" {
		user.AvatarURL = &req.AvatarURL
	}
	if req.Bio != "" {
		user.Bio = &req.Bio
	}

	if err := h.userRepo.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al actualizar perfil"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"role":       user.Role,
			"avatar_url": user.AvatarURL,
			"bio":        user.Bio,
		},
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.authSvc.InvalidateRefreshToken(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al cerrar sesión"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "sesión cerrada"}})
}

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email es requerido"})
		return
	}

	user, err := h.userRepo.FindByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "si el email existe, recibirás un enlace de recuperación"}})
		return
	}

	if err := h.resetRepo.InvalidateExistingTokens(user.Email); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al procesar solicitud"})
		return
	}

	token, expiresAt, err := h.resetRepo.CreateToken(user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al generar token"})
		return
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", c.Request.Header.Get("Origin"), token)

	fmt.Printf("[PASSWORD RESET] Email: %s | Link: %s | Expires: %s\n", user.Email, resetLink, expiresAt)

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "si el email existe, recibirás un enlace de recuperación"}})
}

type resetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token y contraseña son requeridos (mín. 8 caracteres)"})
		return
	}

	email, err := h.resetRepo.ValidateToken(req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token inválido o expirado"})
		return
	}

	user, err := h.userRepo.FindByEmail(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al validar usuario"})
		return
	}

	hash, err := h.authSvc.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al procesar contraseña"})
		return
	}

	user.PasswordHash = hash
	if err := h.userRepo.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al actualizar contraseña"})
		return
	}

	if err := h.resetRepo.MarkTokenUsed(req.Token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al marcar token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "contraseña actualizada correctamente"}})
}
