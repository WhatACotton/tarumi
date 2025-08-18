package models

import "time"

type FirebaseUser struct {
	Email  string
	UserID string
}

type User struct {
	Email    string
	UserID   string
	UserName string
}

type Todo struct {
	ID        string
	Title     string
	Content   string
	StartDate time.Time
	EndDate   time.Time
	CreatedAt time.Time
	ParentID  *string
}
