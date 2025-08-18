package middleware

import (
	"context"
	"errors"
	"log"
	"os"

	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
	"github.com/whatacotton/tarumi/internal/models"
	"google.golang.org/api/option"
)

type FirebaseClaim string

const (
	ClaimEmail           FirebaseClaim = "email"
	ClaimUserId          FirebaseClaim = "user_id"
	ClaimIsEmailVerified FirebaseClaim = "email_verified"
)

type FirebaseService struct {
	app *firebase.App
}

func GetFirebaseService(c *gin.Context) (*FirebaseService, error) {
	app, err := initFireBase()
	if err != nil {
		return nil, err
	}
	s := &FirebaseService{
		app: app,
	}
	return s, nil
}

func initFireBase() (*firebase.App, error) {
	projID := os.Getenv("FIREBASE_PROJECT_ID")
	if projID == "" {
		log.Fatal("FIREBASE_PROJECT_ID is not set in environment variables")
		return nil, nil
	}
	conf := &firebase.Config{
		ProjectID: projID,
	}
	opt := option.WithCredentialsFile("tarumi_credentials.json")
	app, err := firebase.NewApp(context.Background(), conf, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
		return nil, err
	}
	return app, nil
}

// HeaderのAuthenticationに入っているJWTからEmail,UserID,EmailVerifiedを取得
func (s *FirebaseService) GetUser(c *gin.Context) (*models.FirebaseUser, error) {
	user, err := s.getIDToken(c)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *FirebaseService) getIDToken(ctx *gin.Context) (*models.FirebaseUser, error) {
	jwtToken := ctx.Request.Header.Get("Authorization")
	client, err := s.app.Auth(ctx)
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
		return nil, err
	}

	token, err := client.VerifyIDToken(ctx, jwtToken)
	if err != nil {
		log.Fatalf("error verifying ID token: %v\n", err)
		return nil, err
	}
	user := &models.FirebaseUser{
		Email:  token.Claims[string(ClaimEmail)].(string),
		UserID: token.Claims[string(ClaimUserId)].(string),
	}
	if user.Email == "" || user.UserID == "" {
		log.Println("Invalid token: missing email or user_id")
		return nil, errors.New("invalid token: missing email or user_id")
	}
	return user, nil
}
