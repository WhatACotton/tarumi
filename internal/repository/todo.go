package repository

import (
	"time"
	"github.com/whatacotton/tarumi/internal/models"
)

type TodoRepository interface {
	CreateTodo(userID string, title string, description string, startDate time.Time, endDate time.Time) (*models.Todo, error)
	UpdateTodo(userID string, todoID string, title string, description string, startDate time.Time, endDate time.Time) (*models.Todo, error)
	DeleteTodo(userID string, todoID string) error
	GetTodoByID(userID string, todoID string) (*models.Todo, error)
	GetTodosByUserID(userID string) ([]*models.Todo, error)
}