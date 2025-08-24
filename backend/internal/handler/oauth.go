package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/whatacotton/tarumi/internal/middleware"
)

func HandleToken(r *gin.Engine) {
	authHandler := r.Group("/token")
	authHandler.Use(middleware.AuthMiddleware)

	h := OAuthHandler{}
	authHandler.POST("/validate", h.ValidateToken)
}

type OAuthHandler struct{}

type TokenRequest struct {
	AccessToken string `json:"access_token" binding:"required"`
	TokenType   string `json:"token_type"`
}

// ValidateToken はフロントエンドから送信されたアクセストークンの検証を行います
func (h *OAuthHandler) ValidateToken(c *gin.Context) {
	userID, exists := c.Get(string(middleware.ClaimUserId))
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}

	var tokenReq TokenRequest
	if err := c.ShouldBindJSON(&tokenReq); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid token payload"})
		return
	}

	// トークンが提供されているかチェック
	if tokenReq.AccessToken == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Access token is required"})
		return
	}

	// 成功レスポンス
	c.JSON(200, gin.H{
		"message":    "Token validated successfully",
		"user_id":    userID,
		"token_type": tokenReq.TokenType,
		"has_token":  true,
	})
}
