package repository

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/whatacotton/tarumi/internal/config"
	"github.com/whatacotton/tarumi/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMain(m *testing.M) {
	setupTestDB()
	code := m.Run()
	teardownTestDB()
	os.Exit(code)
}

func setupTestDB() {
	// Use test database configuration
	host := getEnvOrDefault("TEST_DB_HOST", "localhost")
	port := getEnvOrDefault("TEST_DB_PORT", "5433") // Changed to test database port
	user := getEnvOrDefault("TEST_DB_USER", "tarumi_user")
	password := getEnvOrDefault("TEST_DB_PASSWORD", "tarumi_password")
	dbname := getEnvOrDefault("TEST_DB_NAME", "tarumi_test")
	sslmode := getEnvOrDefault("TEST_DB_SSLMODE", "disable")

	dsn := "host=" + host + " user=" + user + " password=" + password + " dbname=" + dbname + " port=" + port + " sslmode=" + sslmode + " TimeZone=Asia/Tokyo"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Failed to connect to test database:", err)
	}

	config.DB = db

	// Auto-migrate the schema for testing
	if err := db.AutoMigrate(&models.RepositoryUser{}, &models.RepositoryTodo{}); err != nil {
		log.Fatal("Failed to migrate test database:", err)
	}
}

func teardownTestDB() {
	if config.DB != nil {
		// Clean up test data
		config.DB.Exec("DROP TABLE IF EXISTS repository_todos")
		config.DB.Exec("DROP TABLE IF EXISTS repository_users")

		sqlDB, err := config.DB.DB()
		if err != nil {
			log.Printf("Error getting database instance: %v", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func TestTodoRepository_CreateTodo(t *testing.T) {
	repo := NewTodoRepository()
	userID := "test-user-" + uuid.New().String()

	payload := models.TodoRegisterPayload{
		Title:       "Test Todo",
		Description: "This is a test todo",
		StartDate:   time.Now(),
		EndDate:     time.Now().Add(24 * time.Hour),
		ParentID:    "",
	}

	todo, err := repo.CreateTodo(userID, payload)
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	if todo.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, todo.UserID)
	}
	if todo.Content.Title != payload.Title {
		t.Errorf("Expected Title %s, got %s", payload.Title, todo.Content.Title)
	}
	if todo.Content.Description != payload.Description {
		t.Errorf("Expected Description %s, got %s", payload.Description, todo.Content.Description)
	}
}

func TestTodoRepository_GetTodoByID(t *testing.T) {
	repo := NewTodoRepository()
	userID := "test-user-" + uuid.New().String()

	payload := models.TodoRegisterPayload{
		Title:       "Test Todo",
		Description: "This is a test todo",
		StartDate:   time.Now(),
		EndDate:     time.Now().Add(24 * time.Hour),
		ParentID:    "",
	}

	// Create a todo first
	createdTodo, err := repo.CreateTodo(userID, payload)
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Get the todo by ID
	retrievedTodo, err := repo.GetTodoByID(userID, createdTodo.ID)
	if err != nil {
		t.Fatalf("Failed to get todo by ID: %v", err)
	}

	if retrievedTodo.ID != createdTodo.ID {
		t.Errorf("Expected ID %s, got %s", createdTodo.ID, retrievedTodo.ID)
	}
	if retrievedTodo.Content.Title != payload.Title {
		t.Errorf("Expected Title %s, got %s", payload.Title, retrievedTodo.Content.Title)
	}
}

func TestTodoRepository_UpdateTodo(t *testing.T) {
	repo := NewTodoRepository()
	userID := "test-user-" + uuid.New().String()

	createPayload := models.TodoRegisterPayload{
		Title:       "Test Todo",
		Description: "This is a test todo",
		StartDate:   time.Now(),
		EndDate:     time.Now().Add(24 * time.Hour),
		ParentID:    "",
	}

	// Create a todo first
	createdTodo, err := repo.CreateTodo(userID, createPayload)
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Update the todo
	newTitle := "Updated Todo"
	newDescription := "This is an updated todo"
	isCompleted := false

	updatePayload := models.TodoUpdatePayload{
		Title:       &newTitle,
		Description: &newDescription,
		StartDate:   &createPayload.StartDate,
		EndDate:     &createPayload.EndDate,
		ParentID:    &createPayload.ParentID,
		IsCompleted: &isCompleted,
	}

	updatedTodo, err := repo.UpdateTodo(userID, createdTodo.ID, updatePayload)
	if err != nil {
		t.Fatalf("Failed to update todo: %v", err)
	}

	if updatedTodo.Content.Title != newTitle {
		t.Errorf("Expected Title %s, got %s", newTitle, updatedTodo.Content.Title)
	}
	if updatedTodo.Content.Description != newDescription {
		t.Errorf("Expected Description %s, got %s", newDescription, updatedTodo.Content.Description)
	}
}

func TestTodoRepository_DeleteTodo(t *testing.T) {
	repo := NewTodoRepository()
	userID := "test-user-" + uuid.New().String()

	payload := models.TodoRegisterPayload{
		Title:       "Test Todo",
		Description: "This is a test todo",
		StartDate:   time.Now(),
		EndDate:     time.Now().Add(24 * time.Hour),
		ParentID:    "",
	}

	// Create a todo first
	createdTodo, err := repo.CreateTodo(userID, payload)
	if err != nil {
		t.Fatalf("Failed to create todo: %v", err)
	}

	// Delete the todo
	err = repo.DeleteTodo(userID, createdTodo.ID)
	if err != nil {
		t.Fatalf("Failed to delete todo: %v", err)
	}

	// Try to get the deleted todo (should fail)
	_, err = repo.GetTodoByID(userID, createdTodo.ID)
	if err == nil {
		t.Error("Expected error when getting deleted todo, but got none")
	}
}

func TestTodoRepository_GetTodosByUserID(t *testing.T) {
	repo := NewTodoRepository()
	userID := "test-user-" + uuid.New().String()

	// Create multiple todos
	for i := 0; i < 3; i++ {
		payload := models.TodoRegisterPayload{
			Title:       "Test Todo " + string(rune(i+49)), // 49 is ASCII for '1'
			Description: "This is test todo " + string(rune(i+49)),
			StartDate:   time.Now(),
			EndDate:     time.Now().Add(24 * time.Hour),
			ParentID:    "",
		}

		_, err := repo.CreateTodo(userID, payload)
		if err != nil {
			t.Fatalf("Failed to create todo %d: %v", i+1, err)
		}
	}

	// Get all todos for the user
	todos, err := repo.GetTodosByUserID(userID)
	if err != nil {
		t.Fatalf("Failed to get todos by user ID: %v", err)
	}

	if len(todos) != 3 {
		t.Errorf("Expected 3 todos, got %d", len(todos))
	}

	for _, todo := range todos {
		if todo.UserID != userID {
			t.Errorf("Expected UserID %s, got %s", userID, todo.UserID)
		}
	}
}
