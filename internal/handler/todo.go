package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/whatacotton/tarumi/internal/middleware"
	"github.com/whatacotton/tarumi/internal/models"
	"github.com/whatacotton/tarumi/internal/repository"
)

func HandleTodo(r *gin.Engine) {
	r.Use(middleware.AuthMiddleware)
	authHandler := r.Group("/todo")
	authHandler.Use(middleware.AuthMiddleware)
	h := &TodoHandler{
		repo: repository.NewTodoRepository(),
	}
	authHandler.POST("/create", h.CreateTodo)
	authHandler.POST("/update", h.UpdateTodo)
	authHandler.GET("/complete", h.CompleteTodo)
	authHandler.GET("/delete", h.DeleteTodo)
	authHandler.GET("/get", h.GetTodoByID)
	authHandler.GET("/list", h.GetTodosByUserID)
}

type TodoHandler struct {
	repo repository.TodoRepository
}

func (h *TodoHandler) CreateTodo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}
	todoPayload := models.TodoRegisterPayload{}
	if err := c.ShouldBindJSON(&todoPayload); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid request payload"})
		return
	}
	todo, err := h.repo.CreateTodo(
		userID.(string),
		todoPayload,
	)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to create todo"})
		return
	}
	c.JSON(200, gin.H{"todo": todo})
}

func (h *TodoHandler) UpdateTodo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}
	todoID, exists := c.GetQuery("id")
	if !exists || todoID == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Todo ID not found"})
		return
	}
	todoPayload := models.TodoUpdatePayload{}
	if err := c.ShouldBindJSON(&todoPayload); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid request payload"})
		return
	}
	todo, err := h.repo.UpdateTodo(
		userID.(string),
		todoID,
		todoPayload,
	)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to update todo"})
		return
	}
	c.JSON(200, gin.H{"todo": todo})
}

func (h *TodoHandler) DeleteTodo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}
	todoID, exists := c.GetQuery("id")
	if !exists || todoID == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Todo ID not found"})
		return
	}
	err := h.repo.DeleteTodo(userID.(string), todoID)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to delete todo"})
		return
	}
	c.JSON(200, gin.H{"message": "Todo deleted successfully"})
}

func (h *TodoHandler) GetTodoByID(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}
	todoID, exists := c.GetQuery("id")
	if !exists || todoID == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Todo ID not found"})
		return
	}
	todo, err := h.repo.GetTodoByID(userID.(string), todoID)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to retrieve todo"})
		return
	}
	c.JSON(200, gin.H{"todo": todo})
}

func (h *TodoHandler) GetTodosByUserID(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}
	todos, err := h.repo.GetTodosByUserID(userID.(string))
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to retrieve todos"})
		return
	}
	c.JSON(200, gin.H{"todos": todos})
}

func (r *TodoHandler) CompleteTodo(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}
	todoID, exists := c.GetQuery("id")
	if !exists || todoID == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Todo ID not found"})
		return
	}
	t := true
	p := models.TodoUpdatePayload{
		IsCompleted: &t,
	}
	todo, err := r.repo.UpdateTodo(
		userID.(string),
		todoID,
		p,
	)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to complete todo"})
		return
	}
	c.JSON(200, gin.H{"todo": todo})
}
