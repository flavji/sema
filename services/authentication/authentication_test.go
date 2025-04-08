package authentication_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	firebase "firebase.google.com/go"
	"github.com/stretchr/testify/assert"
	"google.golang.org/api/option"
	"sema/services/authentication"
)

var authService *authentication.AuthService
var testUID string

func TestMain(m *testing.M) {
	os.Setenv("FIREBASE_AUTH_EMULATOR_HOST", "127.0.0.1:9099")
	os.Setenv("GCLOUD_PROJECT", "test-project")

	app, err := firebase.NewApp(context.Background(), &firebase.Config{ProjectID: "test-project"}, option.WithoutAuthentication())
	if err != nil {
		panic(fmt.Sprintf("failed to initialize Firebase App: %v", err))
	}
	authService, err = authentication.NewTestAuthService(app)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize authService: %v", err))
	}

	// Create user via REST API
	reqBody := `{"email": "testuser@example.com", "password": "password123"}`
	resp, err := http.Post(
		"http://127.0.0.1:9099/identitytoolkit.googleapis.com/v1/accounts:signUp?key=fake-api-key",
		"application/json",
		strings.NewReader(reqBody),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create test user (request error): %v", err))
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var respData map[string]interface{}
	_ = json.Unmarshal(body, &respData)

	if resp.StatusCode != http.StatusOK {
		bodyStr := string(body)
		if strings.Contains(bodyStr, "EMAIL_EXISTS") {
			fmt.Println("User already exists, continuing")
			// Log in to fetch UID
			loginBody := `{"email": "testuser@example.com", "password": "password123", "returnSecureToken": true}`
			loginResp, err := http.Post(
				"http://127.0.0.1:9099/identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=fake-api-key",
				"application/json",
				strings.NewReader(loginBody),
			)
			if err != nil {
				panic(fmt.Sprintf("Failed to log in test user: %v", err))
			}
			defer loginResp.Body.Close()
			loginRespBody, _ := io.ReadAll(loginResp.Body)
			_ = json.Unmarshal(loginRespBody, &respData)
		} else {
			panic(fmt.Sprintf("Failed to create test user (status: %d): %s", resp.StatusCode, bodyStr))
		}
	}

	uid, ok := respData["localId"].(string)
	if !ok {
		panic("failed to get UID from signup/login response")
	}
	testUID = uid

	code := m.Run()
	_ = authService.DestroyUser(testUID)
	os.Exit(code)
}

func TestGetUserByUID(t *testing.T) {
	user, err := authService.GetUserByUID(testUID)
	assert.NoError(t, err)
	assert.Equal(t, "testuser@example.com", user.Email)
}

func TestGetUIDFromEmail(t *testing.T) {
	uid, err := authService.GetUIDFromEmail("testuser@example.com")
	assert.NoError(t, err)
	assert.Equal(t, testUID, uid)
}

func TestDestroyUser(t *testing.T) {
	err := authService.DestroyUser(testUID)
	assert.NoError(t, err)

	_, err = authService.GetUserByUID(testUID)
	assert.Error(t, err)
}

func TestVerifyToken_Invalid(t *testing.T) {
	_, err := authService.VerifyToken("invalid.token.here")
	assert.Error(t, err)
}

