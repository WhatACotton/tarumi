package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/whatacotton/tarumi/internal/middleware"
	"github.com/whatacotton/tarumi/internal/models"
	firebase "firebase.google.com/go"
)

func HandleTodo(r *gin.Engine, app *firebase.App) {
	authHandler := r.Group("/todo", func(ctx *gin.Context) {
		fbservice, err := middleware.GetFirebaseService(ctx, app)
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": "Internal Error"})
			return
		}
		ctx.Set("firebaseService", fbservice)
		userID, _, _, err := fbservice.GetUser(ctx)
		if err != nil {
			ctx.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		ctx.Set("userID", userID)
		ctx.Next()
	})
	authHandler.POST("/create", HandleTodoCreate)
}

func HandleTodoCreate(c *gin.Context) {
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

	c.JSON(200, gin.H{"user": userID})
}