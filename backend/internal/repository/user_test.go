package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/whatacotton/tarumi/internal/models"
)

func TestUserRepository_CreateUser(t *testing.T) {
	repo := NewUserRepository()
	userID := "test-user-" + uuid.New().String()
	email := "test-" + uuid.New().String() + "@example.com"
	name := "Test User"

	user, err := repo.GetUser(userID, email, name)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, user.UserID)
	}
	if user.Email != email {
		t.Errorf("Expected Email %s, got %s", email, user.Email)
	}
	if user.UserName != name {
		t.Errorf("Expected UserName %s, got %s", name, user.UserName)
	}
	if user.Status.Grade != models.FreeUser {
		t.Errorf("Expected Grade %s, got %s", models.FreeUser, user.Status.Grade)
	}
	if user.Status.Level != 0 {
		t.Errorf("Expected Level 0, got %d", user.Status.Level)
	}
}

func TestUserRepository_GetUserByID(t *testing.T) {
	repo := NewUserRepository()
	userID := "test-user-" + uuid.New().String()
	email := "test-" + uuid.New().String() + "@example.com"
	name := "Test User"

	// Create a user first
	_, err := repo.GetUser(userID, email, name)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Get the user by ID
	retrievedUser, err := repo.GetUserByID(userID)
	if err != nil {
		t.Fatalf("Failed to get user by ID: %v", err)
	}

	if retrievedUser.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, retrievedUser.UserID)
	}
	if retrievedUser.Email != email {
		t.Errorf("Expected Email %s, got %s", email, retrievedUser.Email)
	}
	if retrievedUser.UserName != name {
		t.Errorf("Expected UserName %s, got %s", name, retrievedUser.UserName)
	}
}

func TestUserRepository_ModifyUserName(t *testing.T) {
	repo := NewUserRepository()
	userID := "test-user-" + uuid.New().String()
	email := "test-" + uuid.New().String() + "@example.com"
	name := "Test User"

	// Create a user first
	_, err := repo.GetUser(userID, email, name)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Modify the user name
	newName := "Updated Test User"
	updatePayload := models.UserUpdatePayload{
		UserName: &newName,
	}

	updatedUser, err := repo.ModifyUserName(userID, updatePayload)
	if err != nil {
		t.Fatalf("Failed to modify user name: %v", err)
	}

	if updatedUser.UserName != newName {
		t.Errorf("Expected UserName %s, got %s", newName, updatedUser.UserName)
	}

	// Verify the change persisted
	retrievedUser, err := repo.GetUserByID(userID)
	if err != nil {
		t.Fatalf("Failed to get user by ID after update: %v", err)
	}

	if retrievedUser.UserName != newName {
		t.Errorf("Expected UserName %s after retrieval, got %s", newName, retrievedUser.UserName)
	}
}
