package repository

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/whatacotton/tarumi/internal/config"
	"github.com/whatacotton/tarumi/internal/models"
	"github.com/whatacotton/tarumi/internal/service"
	"gorm.io/gorm"
)

type TodoRepository interface {
	CreateTodo(userID string, p models.TodoRegisterPayload) (*models.Todo, error)
	UpdateTodo(userID string, todoID string, p models.TodoUpdatePayload) (*models.Todo, error)
	DeleteTodo(userID string, todoID string) error
	GetTodoByID(userID string, todoID string) (*models.Todo, error)
	GetTodosByUserID(userID string) ([]*models.Todo, error)
	GetTodosByGroupID(userID string, groupID string) ([]*models.Todo, error)
	GetCompleteTodos(userID string) ([]*models.Todo, error)
	GetIncompleteTodos(userID string) ([]*models.Todo, error)
}

func NewTodoRepository() TodoRepository {
	return &todoRepository{
		db:                config.DB,
		openrouterService: service.NewOpenRouterService(),
	}
}

type todoRepository struct {
	db                *gorm.DB
	openrouterService *service.OpenRouterService
}

func (r *todoRepository) CreateTodo(userId string, p models.TodoRegisterPayload) (*models.Todo, error) {
	log.Printf("[TodoRepo] CreateTodo called for user: %s, title: '%s'", userId, p.Title)

	// OpenRouterでdurationとジャンルを一括推論
	estimatedDuration, inferredGenre, err := service.InferDurationAndGenreWithOpenRouter(p.Title, p.Description)
	if err != nil {
		log.Printf("[TodoRepo] OpenRouter inference error: %v, using defaults", err)
		estimatedDuration = 45
		inferredGenre = "other"
	} else {
		log.Printf("[TodoRepo] OpenRouter estimated duration: %d minutes, genre: %s", estimatedDuration, inferredGenre)
	}

	todoRepo := &models.RepositoryTodo{
		ID:           uuid.New().String(),
		UserID:       userId,
		Title:        p.Title,
		Description:  p.Description,
		DueDate:      p.DueDate,
		Duration:     estimatedDuration, // AI予測またはスマートフォールバック
		IsCompleted:  false,
		ConsumedTime: 0,
		CreatedAt:    time.Now(),
		ParentID:     p.ParentID,
		GroupID:      p.GroupID,
		Genre:        inferredGenre,
	}

	log.Printf("[TodoRepo] Creating todo with ID: %s, duration: %d", todoRepo.ID, todoRepo.Duration)

	if err := r.db.Create(todoRepo).Error; err != nil {
		log.Printf("[TodoRepo] Database create failed: %v", err)
		return nil, err
	}

	log.Printf("[TodoRepo] Todo created successfully with ID: %s", todoRepo.ID)
	return todoRepo.ConvertToTodo(), nil
}

func (r *todoRepository) UpdateTodo(userID string, todoID string, p models.TodoUpdatePayload) (*models.Todo, error) {
	log.Printf("[TodoRepo] UpdateTodo called for user: %s, todo: %s", userID, todoID)

	var todoRepo models.RepositoryTodo

	if err := r.db.Where("id = ? AND user_id = ?", todoID, userID).First(&todoRepo).Error; err != nil {
		log.Printf("[TodoRepo] Failed to find todo: %v", err)
		return nil, err
	}

	// タイトルまたは説明が変更された場合、durationを再計算
	var shouldRecalculateDuration bool

	if p.Title != nil && *p.Title != todoRepo.Title {
		log.Printf("[TodoRepo] Title changed from '%s' to '%s'", todoRepo.Title, *p.Title)
		todoRepo.Title = *p.Title
		shouldRecalculateDuration = true
	}
	if p.Description != nil && *p.Description != todoRepo.Description {
		log.Printf("[TodoRepo] Description changed, length: %d -> %d", len(todoRepo.Description), len(*p.Description))
		todoRepo.Description = *p.Description
		shouldRecalculateDuration = true
	}

	// Durationとジャンルの再推論 (improved error handling)
	if shouldRecalculateDuration {
		log.Printf("[TodoRepo] Recalculating duration and genre via OpenRouter...")
		estimatedDuration, inferredGenre, err := service.InferDurationAndGenreWithOpenRouter(todoRepo.Title, todoRepo.Description)
		if err == nil {
			log.Printf("[TodoRepo] New estimated duration: %d minutes (was: %d), genre: %s (was: %s)", estimatedDuration, todoRepo.Duration, inferredGenre, todoRepo.Genre)
			todoRepo.Duration = estimatedDuration
			todoRepo.Genre = inferredGenre
		} else {
			// The service layer now handles fallbacks internally, so errors should be rare
			log.Printf("[TodoRepo] Duration/genre recalculation returned error: %v, keeping existing values: duration=%d, genre=%s", err, todoRepo.Duration, todoRepo.Genre)
		}
	}

	if p.IsCompleted != nil {
		log.Printf("[TodoRepo] IsCompleted changed to: %t", *p.IsCompleted)
		todoRepo.IsCompleted = *p.IsCompleted

		// 完了状態に応じてCompletedAtを自動設定
		if *p.IsCompleted {
			// 完了にする場合、CompletedAtが設定されていなければ現在時刻を設定
			if todoRepo.CompletedAt == nil {
				now := time.Now()
				todoRepo.CompletedAt = &now
				log.Printf("[TodoRepo] Auto-setting CompletedAt to: %v", now)
			}
		} else {
			// 未完了にする場合、CompletedAtをnullに設定
			todoRepo.CompletedAt = nil
			log.Printf("[TodoRepo] Clearing CompletedAt (set to null)")
		}
	}
	if p.DueDate != nil {
		log.Printf("[TodoRepo] DueDate changed to: %d", *p.DueDate)
		todoRepo.DueDate = *p.DueDate
	}
	if p.ParentID != nil {
		todoRepo.ParentID = *p.ParentID
	} else {
		todoRepo.ParentID = ""
	}
	if p.GroupID != nil {
		todoRepo.GroupID = *p.GroupID
	} else {
		todoRepo.GroupID = ""
	}
	if p.ConsumedTime != nil {
		log.Printf("[TodoRepo] ConsumedTime changed to: %d minutes", *p.ConsumedTime)
		todoRepo.ConsumedTime = *p.ConsumedTime
	}

	// CompletedAtの明示的な設定（IsCompletedの処理後に実行）
	if p.CompletedAt != nil {
		log.Printf("[TodoRepo] Explicitly setting CompletedAt to: %v", *p.CompletedAt)
		todoRepo.CompletedAt = p.CompletedAt
	}

	if err := r.db.Save(&todoRepo).Error; err != nil {
		log.Printf("[TodoRepo] Database save failed: %v", err)
		return nil, err
	}

	log.Printf("[TodoRepo] Todo updated successfully: %s", todoID)
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

func (r *todoRepository) GetTodosByGroupID(userID string, groupID string) ([]*models.Todo, error) {
	var todoRepos []models.RepositoryTodo

	if err := r.db.Where("user_id = ? AND group_id = ?", userID, groupID).Find(&todoRepos).Error; err != nil {
		return nil, err
	}

	todos := make([]*models.Todo, len(todoRepos))
	for i, todoRepo := range todoRepos {
		todos[i] = todoRepo.ConvertToTodo()
	}

	return todos, nil
}

func (r *todoRepository) GetCompleteTodos(userID string) ([]*models.Todo, error) {
	var todoRepos []models.RepositoryTodo

	if err := r.db.Where("user_id = ? AND is_completed = ?", userID, true).Find(&todoRepos).Error; err != nil {
		return nil, err
	}

	todos := make([]*models.Todo, len(todoRepos))
	for i, todoRepo := range todoRepos {
		todos[i] = todoRepo.ConvertToTodo()
	}

	return todos, nil
}

func (r *todoRepository) GetIncompleteTodos(userID string) ([]*models.Todo, error) {
	var todoRepos []models.RepositoryTodo

	if err := r.db.Where("user_id = ? AND is_completed = ?", userID, false).Find(&todoRepos).Error; err != nil {
		return nil, err
	}

	todos := make([]*models.Todo, len(todoRepos))
	for i, todoRepo := range todoRepos {
		todos[i] = todoRepo.ConvertToTodo()
	}

	return todos, nil
}
