package models

import "time"

type LineTalkRoom struct {
	OwnerID      string
	FriendRoomID string
}

type WaitingFriendRoomsQueue struct {
	OwnerID      string
	IdentifyCode string
	CreatedAt    time.Time
}
