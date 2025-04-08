package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"strings"
	"os"
	"time"

	"sema/repository"
	"sema/models/reportTemplates"

	"github.com/gorilla/websocket"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"sema/api/handlers"
	auth "firebase.google.com/go/auth"


)




type mockAuthService struct {
	verifyTokenFunc func(token string) (*auth.Token, error)
	 GetUIDFromEmailFunc func(email string) (string, error)

}


func (m *mockAuthService) GetUIDFromEmail(email string) (string, error) {
	if m.GetUIDFromEmailFunc != nil {
		return m.GetUIDFromEmailFunc(email)
	}
	return "", nil
}


func (m *mockAuthService) GetUserByUID(uid string) (*auth.UserRecord, error) {
	return nil, nil
}


func (m *mockAuthService) DestroyUser(uid string) error {
	return nil
}



func (m *mockAuthService) VerifyToken(token string) (*auth.Token, error) {
	if m.verifyTokenFunc != nil {
		return m.verifyTokenFunc(token)
	}
	return nil, nil
}



type mockRepo struct{
	BufferLogFunc func(reportID, message, user string)
	getReportFieldTemplateIDFunc func(reportID string) (string, error)
	getTemplateFunc func(templateID string) (*reportTemplates.ReportTemplate, error)
	FetchReportContentFunc func(reportID string) (string, []map[string]interface{}, error)
	FetchLogsForReportFunc func(reportID string) ([]string, error)
	RemoveUserFromReportFunc func(uid, reportID string) error
	RenameReportFunc func(reportID, newName string) error
	DeleteReportFunc func(reportID string) error
	IsAdminInReportFunc func(uid, reportID string) (bool, error)
	FetchReportSectionContentsFunc    func(reportID, section string) (map[string]string, error)
	UpdateReportSectionContentsFunc   func(reportID, section, subsection, content string) error
}

func (m *mockRepo) UpdateReportSectionContents(reportID, section, subsection, content string) error {
	return m.UpdateReportSectionContentsFunc(reportID, section, subsection, content)
}

func (m *mockRepo) FetchReportSectionContents(reportID, section string) (map[string]string, error) {
	return m.FetchReportSectionContentsFunc(reportID, section)
}



func (m *mockRepo) IsAdminInReport(uid, reportID string) (bool, error) {
	if m.IsAdminInReportFunc != nil {
		return m.IsAdminInReportFunc(uid, reportID)
	}
	return false, nil
}

func (m *mockRepo) DeleteReport(reportID string) error {
	if m.DeleteReportFunc != nil {
		return m.DeleteReportFunc(reportID)
	}
	return nil
}

func (m *mockRepo) RenameReport(reportID, reportName string) error {
	if m.RenameReportFunc != nil {
		return m.RenameReportFunc(reportID, reportName)
	}
	return nil
}


func (m *mockRepo) RemoveUserFromReport(uid, reportID string) error {
	if m.RemoveUserFromReportFunc != nil {
		return m.RemoveUserFromReportFunc(uid, reportID)
	}
	return nil
}


func (m *mockRepo) FetchReportContent(reportID string) (string, []map[string]interface{}, error) {
	if m.FetchReportContentFunc != nil {
		return m.FetchReportContentFunc(reportID)
	}
	return "", nil, nil
}



func (m *mockRepo) FetchLogsForReport(reportID string) ([]string, error) {
	if m.FetchLogsForReportFunc != nil {
		return m.FetchLogsForReportFunc(reportID)
	}
	return []string{}, nil
}



func (m *mockRepo) GetUserReportLinks(uid string) ([]repository.Report, error) {
	return []repository.Report{{ReportID: "abc123", ReportTitle: "Test Report"}}, nil
}

func (m *mockRepo) IsUserInReport(uid, reportID string) (bool, error) {
	return true, nil
}



func (m *mockRepo) LinkReportWithUser(uID, reportID string, privilege bool, ownership bool) error {
	return nil
}



func (m *mockRepo) CreateReport(reportName, reportID, templateID, userEmail string) error {
	return nil
}


func (m *mockRepo) DestroyUser(uID string) error {
	return nil
}



func (m *mockRepo) GetReportFieldTemplateID(reportID string) (string, error) {
	return m.getReportFieldTemplateIDFunc(reportID)
}
func (m *mockRepo) GetTemplate(templateID string) (*reportTemplates.ReportTemplate, error) {
	return m.getTemplateFunc(templateID)
}


func (m *mockRepo) BufferLog(reportID, message, user string) {}

func TestHomeHandler(t *testing.T) {
	// Set Gin to Test Mode
	gin.SetMode(gin.TestMode)

	// Set up router and inject context
	router := gin.Default()
	router.LoadHTMLGlob("../../templates/*.html") // make sure this path matches yours

	router.Use(func(c *gin.Context) {
		c.Set("uid", "mockUID123") // inject uid into context
	})

	repo := &mockRepo{}
	router.GET("/", handlers.HomeHandler(repo))

	// Perform request
	req, _ := http.NewRequest("GET", "/", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// Assertions
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "Test Report")
}

func TestAddUserToReport(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuth := &mockAuthService{
		GetUIDFromEmailFunc: func(email string) (string, error) {
			return "mockUID", nil
		},
	}

	mockRepo := &mockRepo{}

	router := gin.Default()
	router.POST("/add", handlers.AddUserToReport(mockAuth, mockRepo))

	body := `{
"email": "test@example.com",
		"reportId": "report123",
		"privilege": true
	}`

	req, _ := http.NewRequest(http.MethodPost, "/add", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "User successfully added")
}


func TestVerifyTokenHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Missing Authorization Header", func(t *testing.T) {
		router := gin.Default()

		mockAuth := &mockAuthService{} // doesn't need to verify anything here
		router.GET("/verify", handlers.VerifyTokenHandler(mockAuth))

		req, _ := http.NewRequest("GET", "/verify", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Authorization token is missing")
	})

	t.Run("Valid Token", func(t *testing.T) {
		router := gin.Default()

		mockAuth := &mockAuthService{
			verifyTokenFunc: func(token string) (*auth.Token, error) {
				return &auth.Token{UID: "mockUID123"}, nil
			},
		}
		router.GET("/verify", handlers.VerifyTokenHandler(mockAuth))

		req, _ := http.NewRequest("GET", "/verify", nil)
		req.Header.Set("Authorization", "Bearer mock.token.here")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "mockUID123")
	})
}


func TestCreateReportHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockRepo{}
	router := gin.Default()

	// Inject uid and email into context
	router.Use(func(c *gin.Context) {
		c.Set("uid", "mockUID123")
		c.Set("email", "user@example.com")
	})

	router.POST("/create", handlers.CreateReportHandler(mockRepo))

	body := `{
"name": "New Report",
		"type": "standard"
	}`

	req, _ := http.NewRequest(http.MethodPost, "/create", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"Success":true`)
	assert.Contains(t, w.Body.String(), `"reportID"`)
}


func TestRegisterHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.Default()
	router.LoadHTMLGlob("../../templates/*.html") // adjust path if needed

	router.GET("/register", handlers.RegisterHandler)

	req, _ := http.NewRequest("GET", "/register", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "<html") // or something unique in register.html
}


func TestReportHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock repo
	mockRepo := &mockRepo{
		// simulate a template with one section and two subsections
		getReportFieldTemplateIDFunc: func(reportID string) (string, error) {
			return "template123", nil
		},
		getTemplateFunc: func(templateID string) (*reportTemplates.ReportTemplate, error) {
			return &reportTemplates.ReportTemplate{
				Sections: []reportTemplates.Section{
					{
						Title:       "Introduction",
						Subsections: []string{"Overview", "Scope"},
					},
				},
			}, nil
		},
	}

	router := gin.Default()
	router.LoadHTMLGlob("../../templates/*.html") // adjust if needed
	router.GET("/report/:reportID", handlers.ReportHandler(mockRepo))

	req, _ := http.NewRequest("GET", "/report/mock123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Introduction")
	assert.Contains(t, w.Body.String(), "Overview")
}


func TestGenerateReportHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)
mockRepo := &mockRepo{}
	// Create dummy PDF output file for testing cleanup
	reportID := "test123"
	reportName := "test-report"
	pdfFileName := reportName + ".pdf"

	// Override the FetchReportContentFunc to return valid content
	mockRepo.FetchReportContentFunc = func(rID string) (string, []map[string]interface{}, error) {
		return reportName, []map[string]interface{}{
			{"sectionTitle": "Introduction", "subsections": []map[string]interface{}{}},
		}, nil
	}

	// Make sure to delete leftover PDF if exists
	_ = os.Remove(pdfFileName)

	router := gin.Default()
	router.GET("/api/generateReport", handlers.GenerateReportHandler(mockRepo))

	req, _ := http.NewRequest(http.MethodGet, "/api/generateReport?reportID="+reportID, nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "application/pdf", resp.Header().Get("Content-Type"))
	assert.Contains(t, resp.Header().Get("Content-Disposition"), pdfFileName)

	// Wait for cleanup goroutine to delete PDF
	time.Sleep(2 * time.Second)

	// Ensure file was cleaned up
	if _, err := os.Stat(pdfFileName); !os.IsNotExist(err) {
		t.Errorf("PDF file %s was not deleted after generation", pdfFileName)
	}
}


func TestReportLogsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockRepo{
		FetchLogsForReportFunc: func(reportID string) ([]string, error) {
			assert.Equal(t, "mockReportID", reportID)
			return []string{"log1", "log2"}, nil
		},
	}

	router := gin.Default()
	router.LoadHTMLGlob("../../templates/*.html") // adjust path if needed
	router.GET("/logs/:reportID", handlers.ReportLogsHandler(mockRepo))

	req, _ := http.NewRequest("GET", "/logs/mockReportID", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "log1")
	assert.Contains(t, w.Body.String(), "log2")
}


func TestRemoveUserFromReport(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAuth := &mockAuthService{
		GetUIDFromEmailFunc: func(email string) (string, error) {
			assert.Equal(t, "test@example.com", email)
			return "mockUID", nil
		},
	}

	mockRepo := &mockRepo{
		RemoveUserFromReportFunc: func(uid, reportID string) error {
			assert.Equal(t, "mockUID", uid)
			assert.Equal(t, "report123", reportID)
			return nil
		},
	}

	router := gin.Default()
	router.POST("/remove/:reportID", handlers.RemoveUserFromReport(mockAuth, mockRepo))

	body := `{"email": "test@example.com"}`
	req, _ := http.NewRequest("POST", "/remove/report123", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "User successfully removed")
}



func TestRenameReport(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockRepo{
		RenameReportFunc: func(reportID, newName string) error {
			if reportID != "mock123" {
				t.Errorf("Expected reportID 'mock123', got '%s'", reportID)
			}
			if newName != "Updated Report" {
				t.Errorf("Expected report name 'Updated Report', got '%s'", newName)
			}
			return nil
		},
	}

	router := gin.Default()
	router.POST("/rename/:reportID", handlers.RenameReport(mockRepo))

	body := `{"reportname": "Updated Report"}`
	req, _ := http.NewRequest(http.MethodPost, "/rename/mock123", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "report was renamed")
}

func TestDeleteReport(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockRepo{
		DeleteReportFunc: func(reportID string) error {
			if reportID != "mock123" {
				t.Errorf("Expected reportID 'mock123', got '%s'", reportID)
			}
			return nil
		},
	}

	router := gin.Default()
	router.DELETE("/delete/:reportID", handlers.DeleteReport(mockRepo))

	req, _ := http.NewRequest(http.MethodDelete, "/delete/mock123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "User successfully removed from report")
}


func TestIsAdminHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockRepo{
		IsAdminInReportFunc: func(uid, reportID string) (bool, error) {
			assert.Equal(t, "mockUID", uid)
			assert.Equal(t, "report456", reportID)
			return true, nil
		},
	}

	router := gin.Default()

	// Middleware to inject uid
	router.Use(func(c *gin.Context) {
		c.Set("uid", "mockUID")
	})

	router.GET("/isadmin/:reportID", handlers.IsAdmin(mockRepo))

	req, _ := http.NewRequest(http.MethodGet, "/isadmin/report456", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"isAdmin":true`)
}

func TestWebSocketHandlerJoinMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mockRepo{
		FetchReportSectionContentsFunc: func(reportID, section string) (map[string]string, error) {
			return map[string]string{"Overview": `{"ops":[{"insert":"Hello world"}]}`}, nil
		},
		BufferLogFunc: func(reportID, message, user string) {
			// You can assert on this if needed
		},
		UpdateReportSectionContentsFunc: func(reportID, section, subsection, content string) error {
			return nil
		},
	}

	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("email", "test@example.com")
	})
	router.GET("/ws/:reportID", handlers.WebSocketHandler(mockRepo))

	server := httptest.NewServer(router)
	defer server.Close()

	// Convert http://127.0.0.1 -> ws://127.0.0.1
	u := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/mockReport"

	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	assert.NoError(t, err)
	defer ws.Close()

	// Send join message
	joinMessage := map[string]interface{}{
		"type":     "join",
		"reportID": "mockReport",
		"section":  "Introduction",
	}
	err = ws.WriteJSON(joinMessage)
	assert.NoError(t, err)

	// Read response (delta)
	var received map[string]interface{}
	err = ws.ReadJSON(&received)
	assert.NoError(t, err)
	assert.NotNil(t, received)
}

func TestWebSocketHandler_Join(t *testing.T) {
  gin.SetMode(gin.TestMode)
  mockRepo := &mockRepo{}
  router := gin.Default()
  router.GET("/ws/:reportID", handlers.WebSocketHandler(mockRepo))

  server := httptest.NewServer(router)
  defer server.Close()

  wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/test-report"
  header := http.Header{"Origin": []string{"http://localhost"}}

  conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
  if err != nil {
    t.Fatalf("Failed to connect to WebSocket: %v", err)
  }
  defer conn.Close()

  msg := `{"type":"join", "reportID":"test-report", "section":"Section 1"}`
  err = conn.WriteMessage(websocket.TextMessage, []byte(msg))
  assert.NoError(t, err)
}

func TestWebSocketHandler_Sync(t *testing.T) {
  gin.SetMode(gin.TestMode)
  mockRepo := &mockRepo{}
  router := gin.Default()
  router.GET("/ws/:reportID", handlers.WebSocketHandler(mockRepo))

  server := httptest.NewServer(router)
  defer server.Close()

  wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/test-report"
  header := http.Header{"Origin": []string{"http://localhost"}}

  conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
  if err != nil {
    t.Fatalf("Failed to connect to WebSocket: %v", err)
  }
  defer conn.Close()

  msg := `{"type":"sync", "reportID":"test-report", "section":"Section 1", "contents":{"editor1":{"ops":[{"insert":"Hello"}]}}}`
  err = conn.WriteMessage(websocket.TextMessage, []byte(msg))
  assert.NoError(t, err)
}

func TestWebSocketHandler_Delta(t *testing.T) {
  gin.SetMode(gin.TestMode)
  mockRepo := &mockRepo{}
  router := gin.Default()
  router.GET("/ws/:reportID", handlers.WebSocketHandler(mockRepo))

  server := httptest.NewServer(router)
  defer server.Close()

  wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/test-report"
  header := http.Header{"Origin": []string{"http://localhost"}}

  conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
  if err != nil {
    t.Fatalf("Failed to connect to WebSocket: %v", err)
  }
  defer conn.Close()

  msg := `{"type":"delta", "ops":[{"insert":"Delta content"}]}`
  err = conn.WriteMessage(websocket.TextMessage, []byte(msg))
  assert.NoError(t, err)
}

func TestWebSocketHandler_UpdateRepo(t *testing.T) {
  gin.SetMode(gin.TestMode)
  mockRepo := &mockRepo{}
  router := gin.Default()
  router.GET("/ws/:reportID", handlers.WebSocketHandler(mockRepo))

  server := httptest.NewServer(router)
  defer server.Close()

  wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/test-report"
  header := http.Header{"Origin": []string{"http://localhost"}}

  conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
  if err != nil {
    t.Fatalf("Failed to connect to WebSocket: %v", err)
  }
  defer conn.Close()

  msg := `{"type":"updateRepo", "reportID":"test-report", "section":"Section 1", "contents":{"editor1":{"ops":[{"insert":"Updated"}]}}}`
  err = conn.WriteMessage(websocket.TextMessage, []byte(msg))
  assert.NoError(t, err)
}

func TestWebSocketHandler_Close(t *testing.T) {
  gin.SetMode(gin.TestMode)
  mockRepo := &mockRepo{}
  router := gin.Default()
  router.GET("/ws/:reportID", handlers.WebSocketHandler(mockRepo))

  server := httptest.NewServer(router)
  defer server.Close()

  wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/test-report"
  header := http.Header{"Origin": []string{"http://localhost"}}

  conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
  if err != nil {
    t.Fatalf("Failed to connect to WebSocket: %v", err)
  }
  defer conn.Close()

  msg := `{"type":"close"}`
  err = conn.WriteMessage(websocket.TextMessage, []byte(msg))
  assert.NoError(t, err)
}

