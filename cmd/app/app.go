package app

import (
	"github.com/gin-gonic/gin"
	"github.com/whatacotton/tarumi/internal/handler"
	"github.com/whatacotton/tarumi/internal/middleware"
)

func Run() {
	r := gin.Default()

	middleware.CORS(r)

	authHandler := r.Group("/user", func(ctx *gin.Context) {
		fbservice, err := middleware.GetFirebaseService(ctx)
		if err != nil {
			ctx.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		ctx.Set("firebaseService", fbservice)
		ctx.Next()
	})
	authHandler.POST("/login", handler.HandleLogin)
	r.Run()
}
