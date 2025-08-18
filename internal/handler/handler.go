package handler

import "github.com/gin-gonic/gin"

func HandleLogin(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Login successful",
	})
}
