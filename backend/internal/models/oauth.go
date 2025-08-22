package models

type OAuthTokenPayload struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresAt    int    `json:"expires_at"`
}

type OAuthToken struct {
	ID           string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	UserID       string `gorm:"not null;index" json:"user_id"`
	AccessToken  string `gorm:"not null" json:"access_token"`
	RefreshToken string `gorm:"not null" json:"refresh_token"`
	TokenType    string `gorm:"not null;default:'Bearer'" json:"token_type"`
	ExpiresAt    int    `gorm:"not null" json:"expires_at"`
}
