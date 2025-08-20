package models

import (
	"fmt"
	"time"
)

type RepositoryUser struct {
	EntryUserID    int `gorm:"primaryKey;autoIncrement"`
	Email          string
	UserID         string
	UserName       string
	RegisteredDate time.Time
	Level          int
	Grade          string
	FriendCode1    string
	FriendCode2    string
	FriendCode3    string
}

type User struct {
	Email       string
	UserID      string
	UserName    string
	Status      UserStatus
	FriendCodes FriendCodes
}
type UserStatus struct {
	RegisteredDate time.Time
	EntryUserID    int
	Level          int
	Grade          UserGrade
}

type FriendCodes struct {
	FriendCode1 string
	FriendCode2 string
	FriendCode3 string
}

type UserGrade string

const (
	FreeUser UserGrade = "free"
	PaidUser UserGrade = "paid"
)

func (user User) ConvertToRepository() *RepositoryUser {
	return &RepositoryUser{
		Email:          user.Email,
		UserID:         user.UserID,
		UserName:       user.UserName,
		RegisteredDate: user.Status.RegisteredDate,
		EntryUserID:    user.Status.EntryUserID,
		Level:          user.Status.Level,
		Grade:          string(user.Status.Grade),
		FriendCode1:    user.FriendCodes.FriendCode1,
		FriendCode2:    user.FriendCodes.FriendCode2,
		FriendCode3:    user.FriendCodes.FriendCode3,
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
			Grade:          grade,
		},
		FriendCodes: FriendCodes{
			FriendCode1: repoUser.FriendCode1,
			FriendCode2: repoUser.FriendCode2,
			FriendCode3: repoUser.FriendCode3,
		},
	}, nil
}
