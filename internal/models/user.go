package models

import (
	"time"
	"fmt"
)

type RepositoryUser struct{
	EntryUserID int  `gorm:"primaryKey;autoIncrement"`
	Email    string 
	UserID   string
	UserName string
	RegisteredDate time.Time
	Level int
	Grade string
}

type User struct {
	Email    string
	UserID   string
	UserName string
	Status UserStatus
}
type UserStatus struct{
	RegisteredDate time.Time
	EntryUserID int
	Level int
	Grade UserGrade
}

type UserGrade string

const (
	 FreeUser UserGrade= "free"
	 PaidUser UserGrade= "paid"
)

func(user User) ConvertToRepository() *RepositoryUser {
	return &RepositoryUser{
		Email:          user.Email,
		UserID:         user.UserID,
		UserName:      user.UserName,
		RegisteredDate: user.Status.RegisteredDate,
		EntryUserID: user.Status.EntryUserID,
		Level: user.Status.Level,
		Grade: string(user.Status.Grade),
	}
}

func (repoUser RepositoryUser) ConvertToUser() (user *User, err error) {
	grade := UserGrade(repoUser.Grade)
	if grade != FreeUser && grade != PaidUser {
		return nil, fmt.Errorf("invalid user grade: %s", grade)
	}
	return &User{
		Email:    repoUser.Email,
		UserID:   repoUser.UserID,
		UserName: repoUser.UserName,
		Status: UserStatus{
			RegisteredDate: repoUser.RegisteredDate,
			EntryUserID:    repoUser.EntryUserID,
			Level:          repoUser.Level,
			Grade:         grade,
		},
	}, nil
}

