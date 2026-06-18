package handlers

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/calli-machtia/backend/internal/middleware"
	"github.com/calli-machtia/backend/internal/repository"
	"github.com/calli-machtia/backend/internal/services"
)

type PaymentHandler struct {
	stripeSvc  *services.StripeService
	paymentRepo *repository.PaymentRepo
}

func NewPaymentHandler(stripeSvc *services.StripeService, paymentRepo *repository.PaymentRepo) *PaymentHandler {
	return &PaymentHandler{stripeSvc: stripeSvc, paymentRepo: paymentRepo}
}

type createIntentRequest struct {
	CourseID string `json:"course_id" binding:"required"`
}

func (h *PaymentHandler) CreateIntent(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req createIntentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "course_id es requerido"})
		return
	}

	paymentIntent, err := h.stripeSvc.CreatePaymentIntent(0, "usd", req.CourseID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al crear intent de pago"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"clientSecret": paymentIntent.ClientSecret,
			"paymentIntentId": paymentIntent.ID,
		},
	})
}

func (h *PaymentHandler) Webhook(c *gin.Context) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "error al leer payload"})
		return
	}

	sigHeader := c.GetHeader("Stripe-Signature")
	if sigHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stripe-signature header requerido"})
		return
	}

	if err := h.stripeSvc.HandleWebhook(payload, sigHeader); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}

func (h *PaymentHandler) History(c *gin.Context) {
	userID := middleware.GetUserID(c)

	payments, err := h.paymentRepo.FindByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error al cargar pagos"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": payments})
}
