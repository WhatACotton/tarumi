package app

import (
	"github.com/gin-gonic/gin"
	"github.com/whatacotton/tarumi/internal/handler"
	"github.com/whatacotton/tarumi/internal/middleware"
)

func Run() {
	r := gin.Default()

	middleware.CORS(r)
	app, err := middleware.InitFireBase()
	if err != nil {
		panic("Failed to initialize Firebase: " + err.Error())
	}
	middleware.FirebaseMiddleware(r, app)
	handler.HandleUser(r)
	handler.HandleTodo(r)
	r.Run()
}
