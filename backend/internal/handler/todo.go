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
	h := &TodoHandler{
		repo: repository.NewTodoRepository(),
	}
	authHandler.Use(middleware.AuthMiddleware)

	authHandler.POST("/create", h.CreateTodo)
	authHandler.POST("/update", h.UpdateTodo)
	authHandler.GET("/complete", h.GetCompleteTodos)
	authHandler.GET("/incomplete", h.GetIncompleteTodos)
	authHandler.POST("/complete", h.CompleteTodo)
	authHandler.POST("/incomplete", h.InCompleteTodo)
	authHandler.GET("/delete", h.DeleteTodo)
	authHandler.GET("/get", h.GetTodoByID)
	authHandler.GET("/list", h.GetTodosByUserID)
	authHandler.GET("/group", h.GetTodosByGroupID)
}

type TodoHandler struct {
	repo repository.TodoRepository
}

func (h *TodoHandler) CreateTodo(c *gin.Context) {
	userID, exists := c.Get(string(middleware.ClaimUserId))
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
	userID, exists := c.Get(string(middleware.ClaimUserId))
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
	userID, exists := c.Get(string(middleware.ClaimUserId))
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
	userID, exists := c.Get(string(middleware.ClaimUserId))
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
	userID, exists := c.Get(string(middleware.ClaimUserId))
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

func (h *TodoHandler) GetTodosByGroupID(c *gin.Context) {
	userID, exists := c.Get(string(middleware.ClaimUserId))
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}
	groupID, exists := c.GetQuery("group_id")
	if !exists || groupID == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Group ID not found"})
		return
	}
	todos, err := h.repo.GetTodosByGroupID(userID.(string), groupID)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to retrieve todos by group"})
		return
	}
	c.JSON(200, gin.H{"todos": todos})
}

func (r *TodoHandler) CompleteTodo(c *gin.Context) {
	userID, exists := c.Get(string(middleware.ClaimUserId))
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}
	todoID, exists := c.GetQuery("id")
	if !exists || todoID == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Todo ID not found"})
		return
	}

	p := models.TodoUpdatePayload{}
	if err := c.ShouldBindJSON(&p); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "Invalid request payload"})
		return
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

func (r *TodoHandler) InCompleteTodo(c *gin.Context) {
	userID, exists := c.Get(string(middleware.ClaimUserId))
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}
	todoID, exists := c.GetQuery("id")
	if !exists || todoID == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "Todo ID not found"})
		return
	}
	t := false
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

func (r *TodoHandler) GetCompleteTodos(c *gin.Context) {
	userID, exists := c.Get(string(middleware.ClaimUserId))
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}
	todos, err := r.repo.GetCompleteTodos(userID.(string))
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to retrieve complete todos"})
		return
	}
	c.JSON(200, gin.H{"todos": todos})
}

func (r *TodoHandler) GetIncompleteTodos(c *gin.Context) {
	userID, exists := c.Get(string(middleware.ClaimUserId))
	if !exists || userID == "" {
		c.AbortWithStatusJSON(500, gin.H{"error": "User ID not found"})
		return
	}
	todos, err := r.repo.GetIncompleteTodos(userID.(string))
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "Failed to retrieve incomplete todos"})
		return
	}
	c.JSON(200, gin.H{"todos": todos})
}
