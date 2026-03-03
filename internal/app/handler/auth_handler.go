package handler

import (
	"net/http"

	"go-gin-project/internal/app/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{service: svc}
}

// Login godoc
// @Summary User login
// @Tags auth
// @Accept json
// @Produce json
// @Param login body service.LoginRequest true "Login credentials"
// @Success 200 {object} service.LoginResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.service.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
