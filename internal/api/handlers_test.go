package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"novabackup/internal/database"
	"novabackup/internal/rbac"

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
	RBACEngine = rbac.NewRBACEngine()
	RBACEngine.DB = db
	_ = RBACEngine.LoadUsersFromDB()

	router := gin.Default()

	// Setup routes
	router.GET("/api/health", GetHealth)
	router.POST("/api/auth/login", Login)

	protected := router.Group("/api")
	protected.Use(AuthMiddleware())
	{
		protected.GET("/jobs", RequirePermission(rbac.PermJobsRead), ListJobs)
		protected.POST("/jobs", RequirePermission(rbac.PermJobsCreate), CreateJob)
	}

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

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%v'", response["status"])
	}
}

func TestLoginSuccess(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()
	defer os.Remove("test_api.db")

	// Create test user
	_, err := RBACEngine.CreateUser("testuser", "MyStr0ngP@ss!", "test@example.com", "", "admin")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	loginReq := map[string]string{
		"username": "testuser",
		"password": "MyStr0ngP@ss!",
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

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestCreateJobEndpoint(t *testing.T) {
	router, db := setupTestRouter(t)
	defer db.Close()
	defer os.Remove("test_api.db")

	// Login to get token
	loginReq := map[string]string{
		"username": "admin",
		"password": "SecurePass1!",
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
		"name":           "Test Job",
		"type":           "file",
		"sources":        []string{"C:\\test"},
		"destination":    "D:\\backup",
		"schedule":       "daily",
		"enabled":        true,
		"retention_days": 30,
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

	// Login
	loginReq := map[string]string{
		"username": "admin",
		"password": "SecurePass1!",
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
