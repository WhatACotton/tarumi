package repository

import (
	"errors"

	"github.com/whatacotton/tarumi/internal/config"
	"github.com/whatacotton/tarumi/internal/models"
	"gorm.io/gorm"
)

type TokenQueueRepository interface {
	Enqueue(token models.RepositoryTokenQueue) error
	Dequeue(userID string) (models.RepositoryTokenQueue, error)
}

type TokenQueueDBRepository struct {
	db *gorm.DB
}

func NewTokenQueueDBRepository() *TokenQueueDBRepository {
	return &TokenQueueDBRepository{
		db: config.DB,
	}
}

// Enqueue: DBにtokenを追加または更新
func (r *TokenQueueDBRepository) Enqueue(token models.RepositoryTokenQueue) error {
	var existing models.RepositoryTokenQueue
	err := r.db.Where("user_id = ?", token.UserID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.Create(&token).Error
	} else if err != nil {
		return err
	}
	// Update the existing token
	existing.Token = token.Token
	// 他に更新したいフィールドがあればここでセット
	return r.db.Save(&existing).Error
}

// Dequeue: userIDの最初のtokenを取得し、削除
var ErrQueueEmpty = errors.New("token queue is empty")

func (r *TokenQueueDBRepository) Dequeue(userID string) (models.RepositoryTokenQueue, error) {
	var token models.RepositoryTokenQueue
	err := r.db.Where("user_id = ?", userID).First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.RepositoryTokenQueue{}, ErrQueueEmpty
		}
		return models.RepositoryTokenQueue{}, err
	}
	if err := r.db.Delete(&token).Error; err != nil {
		return models.RepositoryTokenQueue{}, err
	}
	return token, nil
}

// token値からuser_idを取得する
func (r *TokenQueueDBRepository) GetUserIDByToken(token string) (string, error) {
	var queue models.RepositoryTokenQueue
	err := r.db.Where("token = ?", token).First(&queue).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrQueueEmpty
		}
		return "", err
	}
	return queue.UserID, nil
}
