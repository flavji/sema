package routes_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"sema/api/routes"
	"sema/repository"
	"sema/services/authentication"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.LoadHTMLGlob("../../templates/*")

	auth := &authentication.AuthService{}
	repo := &repository.FirestoreRepository{}

	routes.SetupRoutes(router, auth, repo)
	return router
}

func TestPublicRoutes(t *testing.T) {
	router := setupTestRouter()

	t.Run("GET /register", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/register", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
	})

	t.Run("POST /api/auth/verify", func(t *testing.T) {
		body := `{"token": "mocktoken"}`
		req, _ := http.NewRequest("POST", "/api/auth/verify", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		// Currently returns 401 (unauthorized) due to real service not verifying token
		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})
}

func TestProtectedRoutesWithoutToken(t *testing.T) {
	router := setupTestRouter()

	protectedRoutes := []string{
		"GET /",
		"POST /api/reports",
		"DELETE /api/deleteaccount",
		"GET /report/abc/",
		"GET /report/abc/section/xyz",
		"GET /report/abc/api/isadmin",
		"POST /report/abc/api/addusertoreport",
		"GET /report/abc/api/generateReport",
		"DELETE /report/abc/api/removeuser",
		"POST /report/abc/api/renamereport",
		"DELETE /report/abc/api/deletereport",
		"GET /report/abc/logs",
	}

	for _, route := range protectedRoutes {
		t.Run("No token "+route, func(t *testing.T) {
			parts := strings.SplitN(route, " ", 2)
			method := parts[0]
			path := parts[1]

			req, _ := http.NewRequest(method, path, nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			// NOTE: Middleware currently redirects (302), will update to 401 later
			assert.Equal(t, http.StatusFound, resp.Code)
		})
	}
}
