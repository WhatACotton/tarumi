package handler

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/whatacotton/tarumi/internal/models"
	"github.com/whatacotton/tarumi/internal/repository"
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
		fmt.Println("msg", msg)
		fmt.Println("body", p.Events[0])

		id, err := repository.NewTokenQueueDBRepository().GetUserIDByToken(p.Events[0].Message.Text)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to consume code"})
			return
		}
		msg = fmt.Sprintf("Consumed ID: %s, %s", id, p.Events[0].Source.UserID)
		err = repository.NewUserRepository().AddFriendCode(id, p.Events[0].Source.UserID, 1)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to add friend code"})
			return
		}
		service.NewLineMessagingAPIClient().ReplyMessage(msg, p.Events[0].ReplyToken)

		c.JSON(200, gin.H{"body": p})
	})
}
