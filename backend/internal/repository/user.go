package repository

import (
	"time"

	"github.com/whatacotton/tarumi/internal/config"
	"github.com/whatacotton/tarumi/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	GetUser(userId string, email string, name string) (*models.User, error)
	ModifyUserName(userId string, p models.UserUpdatePayload) (*models.User, error)
	GetUserByID(userID string) (*models.User, error)
}

func NewUserRepository() UserRepository {
	return &userRepository{
		db: config.DB,
	}
}

type userRepository struct {
	db *gorm.DB
}

func (r *userRepository) GetUser(userId string, email string, name string) (*models.User, error) {
	repositoryUser := models.RepositoryUser{
		UserID:         userId,
		Email:          email,
		UserName:       name,
		RegisteredDate: time.Now(),
		Level:          0,
		Grade:          string(models.FreeUser),
	}

	if err := r.db.FirstOrCreate(&repositoryUser, models.RepositoryUser{UserID: userId}).Error; err != nil {
		return nil, err
	}

	return repositoryUser.ConvertToUser()
}

func (r *userRepository) ModifyUserName(userId string, p models.UserUpdatePayload) (*models.User, error) {
	var repositoryUser models.RepositoryUser

	if err := r.db.Where("user_id = ?", userId).First(&repositoryUser).Error; err != nil {
		return nil, err
	}
	if p.UserName == nil || *p.UserName == "" {
		return nil, gorm.ErrInvalidData
	}

	repositoryUser.UserName = *p.UserName
	if err := r.db.Save(&repositoryUser).Error; err != nil {
		return nil, err
	}

	return repositoryUser.ConvertToUser()
}

func (r *userRepository) GetUserByID(userID string) (*models.User, error) {
	var repositoryUser models.RepositoryUser

	if err := r.db.Where("user_id = ?", userID).First(&repositoryUser).Error; err != nil {
		return nil, err
	}

	return repositoryUser.ConvertToUser()
}
