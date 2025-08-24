package repository

import (
	"time"

	"github.com/whatacotton/tarumi/internal/config"
	"github.com/whatacotton/tarumi/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	GetUser(userId string, email string, name string) (*models.User, error)
	UpdateUser(userId string, p models.UserUpdatePayload) (*models.User, error)
	GetUserByID(userID string) (*models.User, error)

	AddFriendCode(userID string, code string, slot int) error
	UpdateFriendCode(userID string, code string, slot int) error
	DeleteFriendCode(userID string, slot int) error
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

// FriendCode追加
func (r *userRepository) AddFriendCode(userID string, code string, slot int) error {
	var user models.RepositoryUser
	if err := r.db.Where("user_id = ?", userID).First(&user).Error; err != nil {
		return err
	}
	switch slot {
	case 1:
		user.FriendCode1 = code
	case 2:
		user.FriendCode2 = code
	case 3:
		user.FriendCode3 = code
	default:
		return gorm.ErrInvalidData
	}
	return r.db.Save(&user).Error
}

// FriendCode修正
func (r *userRepository) UpdateFriendCode(userID string, code string, slot int) error {
	return r.AddFriendCode(userID, code, slot)
}

// FriendCode削除
func (r *userRepository) DeleteFriendCode(userID string, slot int) error {
	var user models.RepositoryUser
	if err := r.db.Where("user_id = ?", userID).First(&user).Error; err != nil {
		return err
	}
	switch slot {
	case 1:
		user.FriendCode1 = ""
	case 2:
		user.FriendCode2 = ""
	case 3:
		user.FriendCode3 = ""
	default:
		return gorm.ErrInvalidData
	}
	return r.db.Save(&user).Error
}

func (r *userRepository) UpdateUser(userId string, p models.UserUpdatePayload) (*models.User, error) {
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
