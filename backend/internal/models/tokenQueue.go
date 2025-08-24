package models

type RepositoryTokenQueue struct {
	ID     uint `gorm:"primaryKey;autoIncrement"`
	Token  string
	UserID string
}
