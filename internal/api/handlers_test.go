package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"novabackup/internal/database"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	code := m.Run()
	os.Exit(code)
}

func setupTestRouter(t *testing.T) (*gin.Engine, *database.Database) {
	dbPath := "test_api.db"
	os.Remove(dbPath)

	db, err := database.NewDatabase(dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	DB = db
	router := gin.Default()

	// Setup routes
	router.POST("/api/auth/login", Login)

	return router, db
}

func TestHealthEndpoint(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()
	defer os.Remove("test_api.db")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%v'", response["status"])
	}
}

func TestLoginSuccess(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()
	defer os.Remove("test_api.db")

	// Create test user
	_, err := db.CreateUser(&database.User{
		Username: "testuser",
		Password: "testpass123", // Note: will be hashed
		Email:    "test@example.com",
		Role:     "admin",
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	loginReq := map[string]string{
		"username": "testuser",
		"password": "testpass123",
	}
	jsonData, _ := json.Marshal(loginReq)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["token"] == "" {
		t.Error("Expected token in response")
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()
	defer os.Remove("test_api.db")

	loginReq := map[string]string{
		"username": "nonexistent",
		"password": "wrongpass",
	}
	jsonData, _ := json.Marshal(loginReq)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestLoginMissingFields(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()
	defer os.Remove("test_api.db")

	loginReq := map[string]string{
		"username": "testuser",
		// password missing
	}
	jsonData, _ := json.Marshal(loginReq)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestCreateJobEndpoint(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()
	defer os.Remove("test_api.db")

	// Create admin user first
	_, err := db.CreateUser(&database.User{
		Username: "admin",
		Password: "admin123",
		Email:    "admin@example.com",
		Role:     "admin",
	})
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	// Login to get token
	loginReq := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	jsonData, _ := json.Marshal(loginReq)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var loginResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResponse)
	token := loginResponse["token"].(string)

	// Create job
	jobReq := map[string]interface{}{
		"name":        "Test Job",
		"type":        "files",
		"sources":     []string{"C:\\test"},
		"destination": "D:\\backup",
		"schedule":    "0 2 * * *",
		"enabled":     true,
	}
	jsonData, _ = json.Marshal(jobReq)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/jobs", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated && w.Code != http.StatusOK {
		t.Errorf("Expected status 200 or 201, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestListJobsEndpoint(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()
	defer os.Remove("test_api.db")

	// Create admin user
	_, err := db.CreateUser(&database.User{
		Username: "admin",
		Password: "admin123",
		Email:    "admin@example.com",
		Role:     "admin",
	})
	if err != nil {
		t.Fatalf("Failed to create admin user: %v", err)
	}

	// Login
	loginReq := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	jsonData, _ := json.Marshal(loginReq)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var loginResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResponse)
	token := loginResponse["token"].(string)

	// List jobs
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/jobs", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["jobs"] == nil {
		t.Error("Expected 'jobs' field in response")
	}
}

func TestUnauthorizedAccess(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()
	defer os.Remove("test_api.db")

	// Try to access protected endpoint without token
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/jobs", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}
