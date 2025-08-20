package models

import "time"

type TodoRegisterPayload struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
}

type TodoUpdatePayload struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
}

type TodoRepository struct {
	ID           string
	UserID       string
	CreatedAt    time.Time
	ParentID     string
	Title        string
	Description  string
	IsCompleted  bool
	StartDate    time.Time
	EndDate      time.Time
	ConsumedTime time.Time
	ExpectedTime time.Time
}

type Todo struct {
	ID        string
	UserID    string
	Content   TodoContent
	CreatedAt time.Time
	ParentID  *string
}

type TodoContent struct {
	Title        string
	Description  string
	IsCompleted  bool
	StartDate    time.Time
	EndDate      time.Time
	ConsumedTime time.Time
	ExpectedTime time.Time
}

type Accomplishments []Accomplishment
type Accomplishment struct {
	Date  time.Time
	Count int
}

func (todo Todo) ConvertToRepository() *TodoRepository {
	if todo.ParentID == nil {
		todo.ParentID = new(string)
	}
	*todo.ParentID = ""
	return &TodoRepository{
		ID:           todo.ID,
		UserID:       todo.UserID,
		CreatedAt:    todo.CreatedAt,
		ParentID:     *todo.ParentID,
		Title:        todo.Content.Title,
		Description:  todo.Content.Description,
		IsCompleted:  todo.Content.IsCompleted,
		StartDate:    todo.Content.StartDate,
		EndDate:      todo.Content.EndDate,
		ConsumedTime: todo.Content.ConsumedTime,
		ExpectedTime: todo.Content.ExpectedTime,
	}
}
func (repoTodo TodoRepository) ConvertToTodo() *Todo {
	var parentID *string
	if repoTodo.ParentID == "" {
		parentID = nil
	} else {
		parentID = &repoTodo.ParentID
	}
	return &Todo{
		ID:        repoTodo.ID,
		UserID:    repoTodo.UserID,
		CreatedAt: repoTodo.CreatedAt,
		ParentID:  parentID,
		Content: TodoContent{
			Title:        repoTodo.Title,
			Description:  repoTodo.Description,
			IsCompleted:  repoTodo.IsCompleted,
			StartDate:    repoTodo.StartDate,
			EndDate:      repoTodo.EndDate,
			ConsumedTime: repoTodo.ConsumedTime,
			ExpectedTime: repoTodo.ExpectedTime,
		},
	}
}
