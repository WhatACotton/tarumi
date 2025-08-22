package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/whatacotton/tarumi/internal/middleware"
	"github.com/whatacotton/tarumi/internal/models"
	"github.com/whatacotton/tarumi/internal/repository"
)

func HandleToken(r *gin.Engine) {
	authHandler := r.Group("/token")
	authHandler.Use(middleware.AuthMiddleware)
	h := OAuthHandler{
		repo: repository.NewOAuthTokenRepository(),
	}
	authHandler.POST("/save", h.SaveToken)
}

type OAuthHandler struct {
	repo repository.OAuthTokenRepository
}

func (h *OAuthHandler) SaveToken(c *gin.Context) {
	userID, exists := c.Get(string(middleware.ClaimUserId))
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}
	eventPayload := models.OAuthTokenPayload{}
	if err := c.ShouldBindJSON(&eventPayload); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid request payload"})
		return
	}
	userID, exists = c.Get(string(middleware.ClaimUserId))
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}
	token := &models.OAuthToken{
		UserID:       userID.(string),
		AccessToken:  eventPayload.AccessToken,
		RefreshToken: eventPayload.RefreshToken,
		TokenType:    eventPayload.TokenType,
		ExpiresAt:    eventPayload.ExpiresAt,
	}
	err := h.repo.Update(token)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to create event"})
		return
	}
	c.JSON(200, gin.H{"event": eventPayload})
}
