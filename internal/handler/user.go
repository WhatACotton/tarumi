package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/whatacotton/tarumi/internal/middleware"
	"github.com/whatacotton/tarumi/internal/repository"
	firebase "firebase.google.com/go"
)

func HandleUser(r *gin.Engine, app *firebase.App) {
	r.POST("/register", func(c *gin.Context) {
		HandleRegister(c, app)
	})
	authHandler := r.Group("/user", func(ctx *gin.Context) {
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
	authHandler.POST("/login", HandleLogin)
}

func HandleLogin(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Login successful",
	})
}


func HandleRegister(c *gin.Context, app *firebase.App) {
	fbservice, err := middleware.GetFirebaseService(c, app)
	if err != nil {
		c.JSON(500, gin.H{"error": "Internal Error"})
		return
	}
	userID, email, name, err := fbservice.GetUser(c)
	if err != nil {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}
	user, err := repository.CreateUser(userID, email, name)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create user"})
		return
	}
	c.JSON(200, gin.H{"user": user})
}