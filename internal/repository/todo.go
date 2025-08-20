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

func NewTodoRepository() TodoRepository {
	return &todoRepository{}
}

type todoRepository struct{}

func (r *todoRepository) CreateTodo(userID string, title string, description string, startDate time.Time, endDate time.Time) (*models.Todo, error) {
	todo := &models.Todo{
		ID:     "mock", // Assume generateID is a function that generates a unique ID
		UserID: userID,
		Content: models.TodoContent{
			Title:        title,
			Description:  description,
			StartDate:    startDate,
			EndDate:      endDate,
			IsCompleted:  false,
			ConsumedTime: time.Time{},
			ExpectedTime: time.Time{},
		},
		CreatedAt: time.Now(),
	}
	// Here you would typically save the todo to a database
	return todo, nil
}
func (r *todoRepository) UpdateTodo(userID string, todoID string, title string, description string, startDate time.Time, endDate time.Time) (*models.Todo, error) {
	// This is a mock implementation. In a real implementation, you would update the todo in the database.
	todo := &models.Todo{
		ID:     todoID,
		UserID: userID,
		Content: models.TodoContent{
			Title:        title,
			Description:  description,
			StartDate:    startDate,
			EndDate:      endDate,
			IsCompleted:  false,
			ConsumedTime: time.Time{},
			ExpectedTime: time.Time{},
		},
		CreatedAt: time.Now(),
	}
	return todo, nil
}

func (r *todoRepository) DeleteTodo(userID string, todoID string) error {
	// This is a mock implementation. In a real implementation, you would delete the todo from the database.
	return nil
}

func (r *todoRepository) GetTodoByID(userID string, todoID string) (*models.Todo, error) {
	// This is a mock implementation. In a real implementation, you would retrieve the todo from the database.
	return &models.Todo{
		ID:     todoID,
		UserID: userID,
		Content: models.TodoContent{
			Title:        "Sample Todo",
			Description:  "This is a sample todo description.",
			StartDate:    time.Now(),
			EndDate:      time.Now().Add(24 * time.Hour),
			IsCompleted:  false,
			ConsumedTime: time.Time{},
			ExpectedTime: time.Time{},
		},
		CreatedAt: time.Now(),
	}, nil
}
func (r *todoRepository) GetTodosByUserID(userID string) ([]*models.Todo, error) {
	// This is a mock implementation. In a real implementation, you would retrieve todos from the database.
	return []*models.Todo{
		{
			ID:     "todo1",
			UserID: userID,
			Content: models.TodoContent{
				Title:        "Sample Todo 1",
				Description:  "This is a sample todo description 1.",
				StartDate:    time.Now(),
				EndDate:      time.Now().Add(24 * time.Hour),
				IsCompleted:  false,
				ConsumedTime: time.Time{},
				ExpectedTime: time.Time{},
			},
			CreatedAt: time.Now(),
		},
	}, nil
}
