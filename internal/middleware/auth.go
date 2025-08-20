package middleware

import (
	"context"
	"errors"
	"log"
	"os"

	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

type FirebaseClaim string

const (
	ClaimEmail           FirebaseClaim = "email"
	ClaimUserId          FirebaseClaim = "user_id"
	ClaimIsEmailVerified FirebaseClaim = "email_verified"
	ClaimDisplayName     FirebaseClaim = "name"
)

type CredentialFilePath string

const (
	TarumiCredentialFile CredentialFilePath = "tarumi_credentials.json"
)

type FirebaseService struct {
	app *firebase.App
}

func FirebaseMiddleware(r *gin.Engine, app *firebase.App) {
	r.Use(func(c *gin.Context) {
		fbservice, err := GetFirebaseService(c, app)
		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "Internal Error"})
			return
		}
		c.Set("firebaseService", fbservice)
		c.Next()
	})
}
func AuthMiddleware(c *gin.Context) {
	fbservice, exists := c.Get("firebaseService")
	if !exists {
		c.AbortWithStatusJSON(500, gin.H{"error": "Firebase service not initialized"})
		return
	}
	service, ok := fbservice.(*FirebaseService)
	if !ok {
		c.AbortWithStatusJSON(500, gin.H{"error": "Invalid Firebase service"})
		return
	}

	userID, email, displayName, err := service.GetUser(c)
	if err != nil {
		c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	c.Set(string(ClaimUserId), userID)
	c.Set(string(ClaimEmail), email)
	c.Set(string(ClaimDisplayName), displayName)
	c.Next()
}

func GetFirebaseService(c *gin.Context, app *firebase.App) (*FirebaseService, error) {
	s := &FirebaseService{
		app: app,
	}
	return s, nil
}

func InitFireBase() (*firebase.App, error) {
	projID := os.Getenv("FIREBASE_PROJECT_ID")
	if projID == "" {
		log.Fatal("FIREBASE_PROJECT_ID is not set in environment variables")
		return nil, nil
	}
	conf := &firebase.Config{
		ProjectID: projID,
	}
	opt := option.WithCredentialsFile(string(TarumiCredentialFile))
	app, err := firebase.NewApp(context.Background(), conf, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
		return nil, err
	}
	return app, nil
}

func (s *FirebaseService) GetUser(c *gin.Context) (userId string, email string, displayName string, err error) {
	jwtToken := c.Request.Header.Get("Authorization")
	if jwtToken == "" {
		log.Fatalf("authorization header is empty")
		return "", "", "", errors.New("authorization header is empty")
	}
	client, err := s.app.Auth(c)
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
		return "", "", "", err
	}
	token, err := client.VerifyIDToken(c, jwtToken)
	if err != nil {
		log.Fatalf("error verifying ID token: %v\n", err)
		return "", "", "", err
	}
	userID := token.Claims[string(ClaimUserId)].(string)
	if userID == "" {
		log.Fatalf("invalid token")
		return "", "", "", errors.New("invalid token")
	}
	email = token.Claims[string(ClaimEmail)].(string)
	if email == "" {
		log.Fatalf("email not found in token claims")
		return "", "", "", errors.New("email not found in token claims")
	}
	displayName = token.Claims[string(ClaimDisplayName)].(string)
	if displayName == "" {
		displayName = "User"
	}
	return userID, email, displayName, nil
}
