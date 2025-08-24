package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/whatacotton/tarumi/internal/lib"
	"github.com/whatacotton/tarumi/internal/middleware"
	"github.com/whatacotton/tarumi/internal/models"
	"github.com/whatacotton/tarumi/internal/repository"
)

func HandleUser(r *gin.Engine) {

	h := &UserHandler{
		repo: repository.NewUserRepository(),
	}
	authHandler := r.Group("/user")
	authHandler.Use(middleware.AuthMiddleware)
	authHandler.POST("/login", HandleLogin)
	authHandler.POST("/rename", h.HandleRename)
	authHandler.GET("/gencode", h.HandleGenerateCode)
}

type UserHandler struct {
	repo repository.UserRepository
}

func HandleLogin(c *gin.Context) {
	userID, exists := c.Get(string(middleware.ClaimUserId))
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}
	email, exists := c.Get(string(middleware.ClaimEmail))
	if !exists || email == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "Email not found"})
		return
	}
	name, exists := c.Get(string(middleware.ClaimDisplayName))
	if !exists || name == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "Display name not found"})
		return
	}

	repo := repository.NewUserRepository()
	user, err := repo.GetUser(userID.(string), email.(string), name.(string))
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create user"})
		return
	}
	c.JSON(200, gin.H{"user": user})
}
func (h *UserHandler) HandleRename(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists || userID == "" {
		c.AbortWithStatusJSON(401, gin.H{"error": "User ID not found"})
		return
	}

	var payload models.UserUpdatePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid request payload"})
		return
	}

	user, err := h.repo.UpdateUser(userID.(string), payload)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to modify user name"})
		return
	}
	c.JSON(200, gin.H{"user": user})
}

func (h *UserHandler) HandleGenerateCode(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists || userID == "" {
		c.AbortWithStatusJSON(401, gin.H{"error": "User ID not found"})
		return
	}

	code, err := lib.CreateNanoid()
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to generate code"})
		return
	}
	fmt.Println("Generated code:", code)
	repository.NewTokenQueueDBRepository().Enqueue(models.RepositoryTokenQueue{
		Token:  code,
		UserID: userID.(string),
	})

	c.JSON(200, gin.H{"code": code})
}
