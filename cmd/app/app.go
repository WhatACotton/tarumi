package app

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/whatacotton/tarumi/internal/config"
	"github.com/whatacotton/tarumi/internal/handler"
	"github.com/whatacotton/tarumi/internal/middleware"
)

func Run() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize database
	config.InitDB()
	defer config.CloseDB()

	r := gin.Default()

	middleware.CORS(r)
	app, err := middleware.InitFireBase()
	if err != nil {
		panic("Failed to initialize Firebase: " + err.Error())
	}
	handler.HandleLineWebhook(r)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	middleware.FirebaseMiddleware(r, app)
	handler.HandleUser(r)
	handler.HandleTodo(r)
	r.Run("0.0.0.0:8080")
}
