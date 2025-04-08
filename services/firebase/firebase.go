package firebase

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

// FirebaseApp holds the initialized Firebase instance
type FirebaseApp struct {
	App *firebase.App
}

// NewFirebaseApp initializes Firebase once
func NewFirebaseApp(credentialsPath string) (*FirebaseApp, error) {
	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing Firebase app: %v", err)
	}

	return &FirebaseApp{App: app}, nil
}
