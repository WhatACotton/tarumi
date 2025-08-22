package repository

import (
	"time"

	"github.com/whatacotton/tarumi/internal/config"
	"github.com/whatacotton/tarumi/internal/models"
	"gorm.io/gorm"
)

type OAuthTokenRepository interface {
	// Create(token *models.OAuthToken) error
	GetByUserID(userID string) (*models.OAuthToken, error)
	Update(token *models.OAuthToken) error
	DeleteByUserID(userID string) error
	IsExpired(userID string) (bool, error)
}

type oauthTokenRepository struct {
	db *gorm.DB
}

func NewOAuthTokenRepository() OAuthTokenRepository {
	return &oauthTokenRepository{db: config.DB}
}

// func (r *oauthTokenRepository) Create(token *models.OAuthToken) error {
// 	return r.db.Create(token).Error
// }

func (r *oauthTokenRepository) GetByUserID(userID string) (*models.OAuthToken, error) {
	var token models.OAuthToken
	err := r.db.Where("user_id = ?", userID).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *oauthTokenRepository) Update(token *models.OAuthToken) error {
	return r.db.Where("user_id = ?", token.UserID).FirstOrCreate(&models.OAuthToken{}, models.OAuthToken{UserID: token.UserID}).Updates(token).Error
}

func (r *oauthTokenRepository) DeleteByUserID(userID string) error {
	return r.db.Where("user_id = ?", userID).Delete(&models.OAuthToken{}).Error
}

func (r *oauthTokenRepository) IsExpired(userID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.OAuthToken{}).
		Where("user_id = ? AND expires_at > ?", userID, time.Now().Unix()).
		Count(&count).Error
	if err != nil {
		return true, err
	}
	return count == 0, nil
}
