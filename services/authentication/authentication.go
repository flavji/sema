package authentication

import (
	"context"
	"fmt"

	"sema/services/firebase"
	testFirebase "firebase.google.com/go"

	"firebase.google.com/go/auth"
)

// AuthClientInterface defines methods for Firebase auth operations
type AuthClientInterface interface {
	VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
	GetUser(ctx context.Context, uid string) (*auth.UserRecord, error)
	GetUserByEmail(ctx context.Context, email string) (*auth.UserRecord, error)
	DeleteUser(ctx context.Context, uid string) error
}

// AuthServiceInterface defines methods used by handlers (for testing and flexibility)
type AuthServiceInterface interface {
	VerifyToken(idToken string) (*auth.Token, error)
	GetUserByUID(uid string) (*auth.UserRecord, error)
	GetUIDFromEmail(email string) (string, error)
	DestroyUser(uid string) error
}

// AuthService handles Firebase Authentication
type AuthService struct {
	AuthClient AuthClientInterface
}




// NewAuthService initializes Firebase Auth using shared FirebaseApp
func NewAuthService(firebaseApp *firebase.FirebaseApp) (*AuthService, error) {
	authClient, err := firebaseApp.App.Auth(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error initializing Firebase Auth: %v", err)
	}

	return &AuthService{AuthClient: authClient}, nil
}

// VerifyToken checks the validity of a Firebase ID token
func (s *AuthService) VerifyToken(idToken string) (*auth.Token, error) {
	return s.AuthClient.VerifyIDToken(context.Background(), idToken)
}

func (s *AuthService) GetUserByUID(uid string) (*auth.UserRecord, error) {
	return s.AuthClient.GetUser(context.Background(), uid)
}

func (s *AuthService) GetUIDFromEmail(email string) (string, error) {
	user, err := s.AuthClient.GetUserByEmail(context.Background(), email)
	if err != nil {
		return "", fmt.Errorf("error fetching user by email: %v", err)
	}
	return user.UID, nil
}

func (s *AuthService) DestroyUser(uid string) error {
	err := s.AuthClient.DeleteUser(context.Background(), uid)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// NewTestAuthService is used only for testing with the Firebase Auth Emulator.
func NewTestAuthService(app *testFirebase.App) (*AuthService, error) {
	authClient, err := app.Auth(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error initializing Firebase Auth: %v", err)
	}
	return &AuthService{AuthClient: authClient}, nil
}
