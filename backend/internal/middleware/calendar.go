package middleware

import (
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/whatacotton/tarumi/internal/repository"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func CalendarMiddleware(c *gin.Context) {
	// Get user ID from context (should be set by auth middleware)
	userID, exists := c.Get(string(ClaimUserId))
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		c.Abort()
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		c.Abort()
		return
	}

	// Get OAuth token from database
	tokenRepo := repository.NewOAuthTokenRepository()
	oauthToken, err := tokenRepo.GetByUserID(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OAuth token not found. Please authenticate with Google Calendar."})
		c.Abort()
		return
	}

	// Check if token is expired
	expired, err := tokenRepo.IsExpired(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check token expiration"})
		c.Abort()
		return
	}

	if expired {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OAuth token has expired. Please re-authenticate."})
		c.Abort()
		return
	}

	// Create OAuth2 token
	token := &oauth2.Token{
		AccessToken:  oauthToken.AccessToken,
		RefreshToken: oauthToken.RefreshToken,
		TokenType:    oauthToken.TokenType,
	}
	// Create OAuth2 config from environment variables
	config := &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
		Scopes:       []string{calendar.CalendarScope},
		Endpoint:     google.Endpoint,
	}

	// Validate OAuth credentials
	if config.ClientID == "" || config.ClientSecret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OAuth credentials not configured"})
		c.Abort()
		return
	}

	// Create HTTP client with OAuth2 token
	client := config.Client(context.Background(), token)

	// Create calendar service
	calendarService, err := calendar.NewService(c.Request.Context(), option.WithHTTPClient(client))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create calendar service"})
		c.Abort()
		return
	}

	// Set calendar service and user ID in context
	c.Set("calendarService", calendarService)
	c.Set("user_id", userIDStr)
	c.Next()

}
