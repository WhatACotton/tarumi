package models

import "time"

type TodoRegisterPayload struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	ParentID    string    `json:"parent_id"`
	GroupID     string    `json:"group_id"`
}

type TodoUpdatePayload struct {
	Title        *string    `json:"title"`
	Description  *string    `json:"description"`
	DueDate      *time.Time `json:"due_date"`
	ParentID     *string    `json:"parent_id"`
	IsCompleted  *bool      `json:"is_completed"`
	GroupID      *string    `json:"group_id"`
	ConsumedTime *int       `json:"consumed_time"` // 分単位
}

type RepositoryTodo struct {
	ID           string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	UserID       string    `gorm:"not null;index"`
	CreatedAt    time.Time `gorm:"not null"`
	ParentID     string    `gorm:"index"`
	GroupID      string    `gorm:"index"`
	Title        string    `gorm:"not null"`
	Description  string
	IsCompleted  bool      `gorm:"default:false"`
	DueDate      time.Time `gorm:"not null"`
	Duration     int       `gorm:"not null"` // 分単位
	ConsumedTime int       // 分単位
}

type Todo struct {
	ID        string
	UserID    string
	Content   TodoContent
	CreatedAt time.Time
	ParentID  *string
	GroupID   *string
}

type TodoContent struct {
	Title        string
	Description  string
	IsCompleted  bool
	DueDate      time.Time
	Duration     int // 分単位
	ConsumedTime int // 分単位
}

type Accomplishments []Accomplishment
type Accomplishment struct {
	Date  time.Time
	Count int
}

func (todo Todo) ConvertToRepository() *RepositoryTodo {
	parentIDValue := ""
	if todo.ParentID != nil {
		parentIDValue = *todo.ParentID
	}

	groupIDValue := ""
	if todo.GroupID != nil {
		groupIDValue = *todo.GroupID
	}

	return &RepositoryTodo{
		ID:           todo.ID,
		UserID:       todo.UserID,
		CreatedAt:    todo.CreatedAt,
		ParentID:     parentIDValue,
		GroupID:      groupIDValue,
		Title:        todo.Content.Title,
		Description:  todo.Content.Description,
		IsCompleted:  todo.Content.IsCompleted,
		DueDate:      todo.Content.DueDate,
		Duration:     todo.Content.Duration,
		ConsumedTime: todo.Content.ConsumedTime,
	}
}

func (repoTodo RepositoryTodo) ConvertToTodo() *Todo {
	var parentID *string
	if repoTodo.ParentID == "" {
		parentID = nil
	} else {
		parentID = &repoTodo.ParentID
	}

	var groupID *string
	if repoTodo.GroupID == "" {
		groupID = nil
	} else {
		groupID = &repoTodo.GroupID
	}

	return &Todo{
		ID:        repoTodo.ID,
		UserID:    repoTodo.UserID,
		CreatedAt: repoTodo.CreatedAt,
		ParentID:  parentID,
		GroupID:   groupID,
		Content: TodoContent{
			Title:        repoTodo.Title,
			Description:  repoTodo.Description,
			IsCompleted:  repoTodo.IsCompleted,
			DueDate:      repoTodo.DueDate,
			Duration:     repoTodo.Duration,
			ConsumedTime: repoTodo.ConsumedTime,
		},
	}
}
