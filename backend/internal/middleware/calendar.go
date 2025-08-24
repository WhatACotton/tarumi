package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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

	// Get OAuth token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "OAuth token not provided. Please include Authorization header."})
		c.Abort()
		return
	}

	// Extract token from "Bearer <token>" format
	tokenParts := strings.SplitN(authHeader, " ", 2)
	if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format. Use 'Bearer <token>'"})
		c.Abort()
		return
	}

	accessToken := tokenParts[1]
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Empty access token"})
		c.Abort()
		return
	}

	// Create OAuth2 token
	token := &oauth2.Token{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(1 * time.Hour), // 仮の有効期限
	}

	// Create OAuth2 config from credentials file or environment
	var clientID, clientSecret string

	// Try to get from environment first
	clientID = os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
	clientSecret = os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET")

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{calendar.CalendarScope, calendar.CalendarEventsScope},
		Endpoint:     google.Endpoint,
	}

	// Create HTTP client with OAuth2 token
	client := config.Client(context.Background(), token)

	// Create calendar service
	calendarService, err := calendar.NewService(c.Request.Context(), option.WithHTTPClient(client))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create calendar service", "details": err.Error()})
		c.Abort()
		return
	}

	// Set calendar service and user ID in context
	c.Set("calendarService", calendarService)
	c.Set("user_id", userIDStr)
	c.Next()
}
