package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/whatacotton/tarumi/internal/config"
	"github.com/whatacotton/tarumi/internal/models"
	"gorm.io/gorm"
)

type TodoRepository interface {
	CreateTodo(userID string, p models.TodoRegisterPayload) (*models.Todo, error)
	UpdateTodo(userID string, todoID string, p models.TodoUpdatePayload) (*models.Todo, error)
	DeleteTodo(userID string, todoID string) error
	GetTodoByID(userID string, todoID string) (*models.Todo, error)
	GetTodosByUserID(userID string) ([]*models.Todo, error)
}

func NewTodoRepository() TodoRepository {
	return &todoRepository{
		db: config.DB,
	}
}

type todoRepository struct {
	db *gorm.DB
}

func (r *todoRepository) CreateTodo(userId string, p models.TodoRegisterPayload) (*models.Todo, error) {
	todoRepo := &models.RepositoryTodo{
		ID:           uuid.New().String(),
		UserID:       userId,
		Title:        p.Title,
		Description:  p.Description,
		StartDate:    p.StartDate,
		EndDate:      p.EndDate,
		IsCompleted:  false,
		ConsumedTime: time.Time{},
		ExpectedTime: time.Time{},
		CreatedAt:    time.Now(),
		ParentID:     p.ParentID,
	}

	if err := r.db.Create(todoRepo).Error; err != nil {
		return nil, err
	}

	return todoRepo.ConvertToTodo(), nil
}

func (r *todoRepository) UpdateTodo(userID string, todoID string, p models.TodoUpdatePayload) (*models.Todo, error) {
	var todoRepo models.RepositoryTodo

	if err := r.db.Where("id = ? AND user_id = ?", todoID, userID).First(&todoRepo).Error; err != nil {
		return nil, err
	}
	if p.IsCompleted != nil {
		todoRepo.IsCompleted = *p.IsCompleted
	}
	if p.Title != nil {
		todoRepo.Title = *p.Title
	}
	if p.Description != nil {
		todoRepo.Description = *p.Description
	}
	if p.StartDate != nil {
		todoRepo.StartDate = *p.StartDate
	}
	if p.EndDate != nil {
		todoRepo.EndDate = *p.EndDate
	}
	if p.ParentID != nil {
		todoRepo.ParentID = *p.ParentID
	} else {
		todoRepo.ParentID = ""
	}

	if err := r.db.Save(&todoRepo).Error; err != nil {
		return nil, err
	}

	return todoRepo.ConvertToTodo(), nil
}

func (r *todoRepository) DeleteTodo(userID string, todoID string) error {
	result := r.db.Where("id = ? AND user_id = ?", todoID, userID).Delete(&models.RepositoryTodo{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *todoRepository) GetTodoByID(userID string, todoID string) (*models.Todo, error) {
	var todoRepo models.RepositoryTodo

	if err := r.db.Where("id = ? AND user_id = ?", todoID, userID).First(&todoRepo).Error; err != nil {
		return nil, err
	}

	return todoRepo.ConvertToTodo(), nil
}
func (r *todoRepository) GetTodosByUserID(userID string) ([]*models.Todo, error) {
	var todoRepos []models.RepositoryTodo

	if err := r.db.Where("user_id = ?", userID).Find(&todoRepos).Error; err != nil {
		return nil, err
	}

	todos := make([]*models.Todo, len(todoRepos))
	for i, todoRepo := range todoRepos {
		todos[i] = todoRepo.ConvertToTodo()
	}

	return todos, nil
}
