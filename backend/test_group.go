package main

import (
	"fmt"
	"time"

	"github.com/whatacotton/tarumi/internal/models"
)

func main() {
	// Test group functionality
	fmt.Println("Testing Todo with Group functionality:")

	// Create sample todo register payload with group_id
	payload := models.TodoRegisterPayload{
		Title:       "Sample Todo",
		Description: "This is a test todo with group",
		StartDate:   time.Now(),
		EndDate:     time.Now().AddDate(0, 0, 7), // 1 week from now
		ParentID:    "",
		GroupID:     "group-123",
	}

	fmt.Printf("TodoRegisterPayload: %+v\n", payload)

	// Create sample todo update payload with group_id
	groupID := "updated-group-456"
	updatePayload := models.TodoUpdatePayload{
		GroupID: &groupID,
	}

	fmt.Printf("TodoUpdatePayload: %+v\n", updatePayload)

	// Create sample repository todo
	repoTodo := models.RepositoryTodo{
		ID:          "test-id",
		UserID:      "user-123",
		CreatedAt:   time.Now(),
		ParentID:    "",
		GroupID:     "group-123",
		Title:       "Sample Todo",
		Description: "This is a test todo with group",
		IsCompleted: false,
		StartDate:   time.Now(),
		EndDate:     time.Now().AddDate(0, 0, 7),
	}

	// Convert to Todo
	todo := repoTodo.ConvertToTodo()
	fmt.Printf("Converted Todo: ID=%s, GroupID=%v\n", todo.ID, todo.GroupID)

	// Convert back to Repository
	backToRepo := todo.ConvertToRepository()
	fmt.Printf("Back to Repository: ID=%s, GroupID=%s\n", backToRepo.ID, backToRepo.GroupID)

	fmt.Println("All tests passed!")
}
