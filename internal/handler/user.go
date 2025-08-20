package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/whatacotton/tarumi/internal/repository"
)

func HandleUser(r *gin.Engine) {
	r.POST("/register", func(c *gin.Context) {
		HandleRegister(c)
	})
	authHandler := r.Group("/user")
	authHandler.POST("/login", HandleLogin)
}

func HandleLogin(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Login successful",
	})
}

func HandleRegister(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}
	email, exists := c.Get("email")
	if !exists || email == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "Email not found"})
		return
	}
	name, exists := c.Get("displayName")
	if !exists || name == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "Display name not found"})
		return
	}
	user, err := repository.CreateUser(userID.(string), email.(string), name.(string))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create user"})
		return
	}
	c.JSON(200, gin.H{"user": user})
}
