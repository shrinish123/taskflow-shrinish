package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shrinishv/taskflow-shrinish/backend/internal/handler"
	"github.com/shrinishv/taskflow-shrinish/backend/internal/model"
	"github.com/shrinishv/taskflow-shrinish/backend/internal/repository"
	"github.com/shrinishv/taskflow-shrinish/backend/internal/router"
	"github.com/shrinishv/taskflow-shrinish/backend/internal/service"
)

// setupTestRouter creates a router wired with real repositories against testPool.
// testPool must be set up before tests run (see TestMain or skip with build tag).
func setupTestRouter(t *testing.T) http.Handler {
	t.Helper()

	if testPool == nil {
		t.Skip("DATABASE_URL not set — skipping integration test")
	}

	userRepo := repository.NewUserRepository(testPool)
	projectRepo := repository.NewProjectRepository(testPool)
	taskRepo := repository.NewTaskRepository(testPool)

	authSvc := service.NewAuthService(userRepo, "test-jwt-secret-for-integration")
	projectSvc := service.NewProjectService(projectRepo, taskRepo)
	taskSvc := service.NewTaskService(taskRepo, projectRepo, userRepo)

	authH := handler.NewAuthHandler(authSvc)
	projectH := handler.NewProjectHandler(projectSvc)
	taskH := handler.NewTaskHandler(taskSvc)

	return router.New(authSvc, authH, projectH, taskH, []string{"http://localhost:3000"})
}

func TestRegisterAndLogin(t *testing.T) {
	r := setupTestRouter(t)

	// Register
	body, _ := json.Marshal(model.RegisterRequest{
		Name:     "Integration User",
		Email:    "integ@test.com",
		Password: "securepass123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("register: expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var regResp model.LoginResponse
	json.NewDecoder(w.Body).Decode(&regResp)
	if regResp.Token == "" {
		t.Fatal("register: expected token in response")
	}
	if regResp.User.Email != "integ@test.com" {
		t.Fatalf("register: expected email integ@test.com, got %s", regResp.User.Email)
	}

	// Login
	loginBody, _ := json.Marshal(model.LoginRequest{
		Email:    "integ@test.com",
		Password: "securepass123",
	})
	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(loginBody))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("login: expected 200, got %d: %s", w2.Code, w2.Body.String())
	}

	var loginResp model.LoginResponse
	json.NewDecoder(w2.Body).Decode(&loginResp)
	if loginResp.Token == "" {
		t.Fatal("login: expected token in response")
	}
}

func TestRegisterValidation(t *testing.T) {
	r := setupTestRouter(t)

	body, _ := json.Marshal(map[string]string{
		"name":  "",
		"email": "not-an-email",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("validation: expected 400, got %d", w.Code)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["error"] != "validation_failed" {
		t.Fatalf("validation: expected error=validation_failed, got %v", resp["error"])
	}
}

func TestProtectedRouteWithoutToken(t *testing.T) {
	r := setupTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("protected: expected 401, got %d", w.Code)
	}
}
