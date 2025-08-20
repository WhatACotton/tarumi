package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/whatacotton/tarumi/internal/lib"
	"github.com/whatacotton/tarumi/internal/models"
	"github.com/whatacotton/tarumi/internal/service"
)

func HandleLineWebhook(r *gin.Engine) {
	r.POST("/webhook/line", func(c *gin.Context) {
		// Handle LINE webhook events
		// debug print bodyPayload
		p := models.LineWebhookEvent{}
		if err := c.ShouldBindJSON(&p); err != nil {
			c.JSON(400, gin.H{"error": "Failed to parse request body"})
			return
		}
		//todo: Eventsが本当に1つだけか確認する
		msg := p.Events[0].Message.Text
		fmt.Println("body", msg)
		if msg == "generate code" {
			id, err := lib.CreateNanoid()
			if err != nil {
				c.JSON(500, gin.H{"error": "Failed to generate ID"})
				return
			}
			msg = fmt.Sprintf("Generated ID: %s", id)
		}

		service.NewLineMessagingAPIClient().ReplyMessage(msg, p.Events[0].ReplyToken)
		c.JSON(200, gin.H{"body": p})
	})
}
