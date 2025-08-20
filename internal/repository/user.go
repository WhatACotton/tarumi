package repository

import (
	"time"
	"github.com/whatacotton/tarumi/internal/models"
)

type UserRepository interface {
	CreateUser(userId string, email string, name string) (*models.User, error)
	ModifyUserName(userId string, name string) (*models.User, error)
}

func CreateUser(userID string, email string, name string) (*models.User, error) {
	repositoryUser := models.RepositoryUser{
		UserID:         userID,
		Email:          email,
		UserName:       name,
		RegisteredDate: time.Now(),
		Level:          0,
		Grade:          string(models.FreeUser),
	}
	// if err := db.Create(&repositoryUser).Error; err != nil {
	// 	return nil, err
	// }
	return repositoryUser.ConvertToUser()
}