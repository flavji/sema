package authentication


import (
	"context"
	"fmt"
	"log"
	

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)



func InitializeAuthenticationClient() *auth.Client {

	// Load Firebase credentials
	opt := option.WithCredentialsFile("../../config/sema-7c193-firebase-adminsdk-fbsvc-4c0f14eca3.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Error initializing Firebase: %v", err)
	}

	// Get Auth client
	authClient, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("Error getting Auth client: %v", err)
	}

	fmt.Println("Firebase initialized successfully")

	return authClient
}


