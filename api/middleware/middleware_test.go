package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"sema/api/middleware"
	"sema/repository"
	"sema/services/authentication"
)

// --- Dummy AuthClientInterface Implementation ---

type dummyAuthClient struct{}

func (d *dummyAuthClient) VerifyIDToken(_ context.Context, token string) (*auth.Token, error) {
	if token == "valid-token" {
		return &auth.Token{UID: "user123"}, nil
	}
	return nil, errors.New("invalid token")
}

func (d *dummyAuthClient) GetUser(_ context.Context, uid string) (*auth.UserRecord, error) {
	if uid == "user123" {
		return &auth.UserRecord{
			UserInfo: &auth.UserInfo{
				UID:   "user123",
				Email: "test@example.com",
			},
		}, nil
	}
	return nil, errors.New("not found")
}

func (d *dummyAuthClient) GetUserByEmail(_ context.Context, email string) (*auth.UserRecord, error) {
	return nil, nil
}

func (d *dummyAuthClient) DeleteUser(_ context.Context, uid string) error {
	return nil
}

// --- Dummy FirestoreRepository Implementation ---

type dummyRepo struct {
	repository.FirestoreRepository
}

func (r *dummyRepo) IsUserInReport(uid, reportID string) (bool, error) {
	return uid == "user123" && reportID == "r1", nil
}

func (r *dummyRepo) IsAdminInReport(uid, reportID string) (bool, error) {
	return uid == "user123" && reportID == "r1", nil
}

// --- Helper ---

func performRequestWithCookie(router *gin.Engine, method, path, cookieValue string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	if cookieValue != "" {
		req.AddCookie(&http.Cookie{Name: "firebaseToken", Value: cookieValue})
	}
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

// --- Tests ---

func TestAuthMiddleware_ValidToken(t *testing.T) {
	auth := &authentication.AuthService{AuthClient: &dummyAuthClient{}}
	router := gin.New()
	router.Use(middleware.AuthMiddleware(auth))
	router.GET("/", func(c *gin.Context) {
		uid, _ := c.Get("uid")
		email, _ := c.Get("email")
		c.JSON(http.StatusOK, gin.H{"uid": uid, "email": email})
	})
	resp := performRequestWithCookie(router, "GET", "/", "valid-token")
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "user123")
	assert.Contains(t, resp.Body.String(), "test@example.com")
}

func TestAuthUserinReport_UserInReport(t *testing.T) {
	auth := &authentication.AuthService{AuthClient: &dummyAuthClient{}}
	repo := &dummyRepo{}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("uid", "user123")
		c.Params = []gin.Param{{Key: "reportID", Value: "r1"}}
	})
	router.Use(middleware.AuthUserinReport(auth, repo))
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "passed")
	})
	req, _ := http.NewRequest("GET", "/", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestAuthAdmininReport_IsAdmin(t *testing.T) {
	auth := &authentication.AuthService{AuthClient: &dummyAuthClient{}}
	repo := &dummyRepo{}

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("uid", "user123")
		c.Params = []gin.Param{{Key: "reportID", Value: "r1"}}
	})
	router.Use(middleware.AuthAdmininReport(auth, repo))
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "passed")
	})
	req, _ := http.NewRequest("GET", "/", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusOK, resp.Code)
}
