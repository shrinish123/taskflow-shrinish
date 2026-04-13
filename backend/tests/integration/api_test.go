package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shrinishv/taskflow-shrinish/backend/internal/model"
)

// ---------- Helpers ----------

func doRequest(t *testing.T, r http.Handler, method, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()
	var reqBody *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewReader(b)
	} else {
		reqBody = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func registerUser(t *testing.T, r http.Handler, name, email, password string) string {
	t.Helper()
	w := doRequest(t, r, http.MethodPost, "/api/auth/register", model.RegisterRequest{
		Name:     name,
		Email:    email,
		Password: password,
	}, "")

	if w.Code != http.StatusCreated {
		t.Fatalf("registerUser(%s): expected 201, got %d: %s", email, w.Code, w.Body.String())
	}

	var resp model.LoginResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("registerUser(%s): decode error: %v", email, err)
	}
	return resp.Token
}

func createProject(t *testing.T, r http.Handler, token, name, desc string) string {
	t.Helper()
	w := doRequest(t, r, http.MethodPost, "/api/projects", model.CreateProjectRequest{
		Name:        name,
		Description: desc,
	}, token)

	if w.Code != http.StatusCreated {
		t.Fatalf("createProject(%s): expected 201, got %d: %s", name, w.Code, w.Body.String())
	}

	var project model.Project
	if err := json.NewDecoder(w.Body).Decode(&project); err != nil {
		t.Fatalf("createProject(%s): decode error: %v", name, err)
	}
	return project.ID
}

func createTask(t *testing.T, r http.Handler, token, projectID, title, status, priority string) string {
	t.Helper()
	w := doRequest(t, r, http.MethodPost, fmt.Sprintf("/api/projects/%s/tasks", projectID), model.CreateTaskRequest{
		Title:    title,
		Status:   status,
		Priority: priority,
	}, token)

	if w.Code != http.StatusCreated {
		t.Fatalf("createTask(%s): expected 201, got %d: %s", title, w.Code, w.Body.String())
	}

	var task model.Task
	if err := json.NewDecoder(w.Body).Decode(&task); err != nil {
		t.Fatalf("createTask(%s): decode error: %v", title, err)
	}
	return task.ID
}

// ---------- Project Tests ----------

func TestCreateProject(t *testing.T) {
	r := setupTestRouter(t)
	token := registerUser(t, r, "Create Project User", "testcreateproject@test.com", "password123")

	w := doRequest(t, r, http.MethodPost, "/api/projects", model.CreateProjectRequest{
		Name:        "My Project",
		Description: "A test project",
	}, token)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var project model.Project
	json.NewDecoder(w.Body).Decode(&project)
	if project.ID == "" {
		t.Fatal("expected project ID")
	}
	if project.Name != "My Project" {
		t.Fatalf("expected name 'My Project', got %q", project.Name)
	}
}

func TestListProjects(t *testing.T) {
	r := setupTestRouter(t)
	token := registerUser(t, r, "List Projects User", "testlistprojects@test.com", "password123")

	createProject(t, r, token, "List Test Project", "desc")

	w := doRequest(t, r, http.MethodGet, "/api/projects", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Projects []model.Project `json:"projects"`
		Total    int             `json:"total"`
	}
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp.Projects) == 0 {
		t.Fatal("expected at least one project")
	}

	found := false
	for _, p := range resp.Projects {
		if p.Name == "List Test Project" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("created project not found in list")
	}
}

func TestGetProject(t *testing.T) {
	r := setupTestRouter(t)
	token := registerUser(t, r, "Get Project User", "testgetproject@test.com", "password123")
	projectID := createProject(t, r, token, "Get Test Project", "some description")

	w := doRequest(t, r, http.MethodGet, fmt.Sprintf("/api/projects/%s", projectID), nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var project model.ProjectWithTasks
	json.NewDecoder(w.Body).Decode(&project)
	if project.Name != "Get Test Project" {
		t.Fatalf("expected name 'Get Test Project', got %q", project.Name)
	}
	if project.Tasks == nil {
		t.Fatal("expected tasks array (even if empty)")
	}
	if len(project.Tasks) != 0 {
		t.Fatalf("expected empty tasks array, got %d tasks", len(project.Tasks))
	}
}

func TestUpdateProject(t *testing.T) {
	r := setupTestRouter(t)
	token := registerUser(t, r, "Update Project User", "testupdateproject@test.com", "password123")
	projectID := createProject(t, r, token, "Original Name", "original desc")

	newName := "Updated Name"
	newDesc := "updated desc"
	w := doRequest(t, r, http.MethodPatch, fmt.Sprintf("/api/projects/%s", projectID), model.UpdateProjectRequest{
		Name:        &newName,
		Description: &newDesc,
	}, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var project model.Project
	json.NewDecoder(w.Body).Decode(&project)
	if project.Name != "Updated Name" {
		t.Fatalf("expected name 'Updated Name', got %q", project.Name)
	}
	if project.Description != "updated desc" {
		t.Fatalf("expected description 'updated desc', got %q", project.Description)
	}
}

func TestUpdateProjectForbidden(t *testing.T) {
	r := setupTestRouter(t)
	ownerToken := registerUser(t, r, "Owner User", "testupdateforbiddenowner@test.com", "password123")
	otherToken := registerUser(t, r, "Other User", "testupdateforbiddenother@test.com", "password123")

	projectID := createProject(t, r, ownerToken, "Owner Project", "desc")

	newName := "Hacked Name"
	w := doRequest(t, r, http.MethodPatch, fmt.Sprintf("/api/projects/%s", projectID), model.UpdateProjectRequest{
		Name: &newName,
	}, otherToken)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteProject(t *testing.T) {
	r := setupTestRouter(t)
	token := registerUser(t, r, "Delete Project User", "testdeleteproject@test.com", "password123")
	projectID := createProject(t, r, token, "To Delete", "will be deleted")

	w := doRequest(t, r, http.MethodDelete, fmt.Sprintf("/api/projects/%s", projectID), nil, token)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}

	// Verify it's gone
	w2 := doRequest(t, r, http.MethodGet, fmt.Sprintf("/api/projects/%s", projectID), nil, token)
	if w2.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d: %s", w2.Code, w2.Body.String())
	}
}

func TestProjectNotFound(t *testing.T) {
	r := setupTestRouter(t)
	token := registerUser(t, r, "NotFound User", "testprojectnotfound@test.com", "password123")

	w := doRequest(t, r, http.MethodGet, "/api/projects/00000000-0000-0000-0000-000000000000", nil, token)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestProjectStats(t *testing.T) {
	r := setupTestRouter(t)
	token := registerUser(t, r, "Stats User", "testprojectstats@test.com", "password123")
	projectID := createProject(t, r, token, "Stats Project", "for stats")

	createTask(t, r, token, projectID, "Task 1", "todo", "low")
	createTask(t, r, token, projectID, "Task 2", "todo", "medium")
	createTask(t, r, token, projectID, "Task 3", "done", "high")

	w := doRequest(t, r, http.MethodGet, fmt.Sprintf("/api/projects/%s/stats", projectID), nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var stats model.ProjectStats
	json.NewDecoder(w.Body).Decode(&stats)
	if stats.TotalTasks != 3 {
		t.Fatalf("expected total_tasks=3, got %d", stats.TotalTasks)
	}
	if stats.ByStatus["todo"] != 2 {
		t.Fatalf("expected by_status[todo]=2, got %d", stats.ByStatus["todo"])
	}
	if stats.ByStatus["done"] != 1 {
		t.Fatalf("expected by_status[done]=1, got %d", stats.ByStatus["done"])
	}
}

// ---------- Task Tests ----------

func TestCreateTask(t *testing.T) {
	r := setupTestRouter(t)
	token := registerUser(t, r, "Create Task User", "testcreatetask@test.com", "password123")
	projectID := createProject(t, r, token, "Task Project", "desc")

	w := doRequest(t, r, http.MethodPost, fmt.Sprintf("/api/projects/%s/tasks", projectID), model.CreateTaskRequest{
		Title:    "My Task",
		Status:   "todo",
		Priority: "high",
	}, token)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var task model.Task
	json.NewDecoder(w.Body).Decode(&task)
	if task.ID == "" {
		t.Fatal("expected task ID")
	}
	if task.Title != "My Task" {
		t.Fatalf("expected title 'My Task', got %q", task.Title)
	}
	if task.Status != "todo" {
		t.Fatalf("expected status 'todo', got %q", task.Status)
	}
	if task.Priority != "high" {
		t.Fatalf("expected priority 'high', got %q", task.Priority)
	}
}

func TestListTasks(t *testing.T) {
	r := setupTestRouter(t)
	token := registerUser(t, r, "List Tasks User", "testlisttasks@test.com", "password123")
	projectID := createProject(t, r, token, "List Tasks Project", "desc")

	createTask(t, r, token, projectID, "Task A", "todo", "low")
	createTask(t, r, token, projectID, "Task B", "in_progress", "medium")
	createTask(t, r, token, projectID, "Task C", "done", "high")

	w := doRequest(t, r, http.MethodGet, fmt.Sprintf("/api/projects/%s/tasks", projectID), nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Tasks []model.Task `json:"tasks"`
		Total int          `json:"total"`
	}
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp.Tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(resp.Tasks))
	}
}

func TestListTasksFilterByStatus(t *testing.T) {
	r := setupTestRouter(t)
	token := registerUser(t, r, "Filter Tasks User", "testfiltertasks@test.com", "password123")
	projectID := createProject(t, r, token, "Filter Tasks Project", "desc")

	createTask(t, r, token, projectID, "Todo Task 1", "todo", "low")
	createTask(t, r, token, projectID, "Todo Task 2", "todo", "medium")
	createTask(t, r, token, projectID, "Done Task", "done", "high")

	w := doRequest(t, r, http.MethodGet, fmt.Sprintf("/api/projects/%s/tasks?status=todo", projectID), nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Tasks []model.Task `json:"tasks"`
		Total int          `json:"total"`
	}
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp.Tasks) != 2 {
		t.Fatalf("expected 2 todo tasks, got %d", len(resp.Tasks))
	}
	for _, task := range resp.Tasks {
		if task.Status != "todo" {
			t.Fatalf("expected all tasks to have status 'todo', got %q", task.Status)
		}
	}
}

func TestUpdateTask(t *testing.T) {
	r := setupTestRouter(t)
	token := registerUser(t, r, "Update Task User", "testupdatetask@test.com", "password123")
	projectID := createProject(t, r, token, "Update Task Project", "desc")
	taskID := createTask(t, r, token, projectID, "Original Task", "todo", "low")

	newStatus := "done"
	newTitle := "Updated Task"
	w := doRequest(t, r, http.MethodPatch, fmt.Sprintf("/api/tasks/%s", taskID), model.UpdateTaskRequest{
		Status: &newStatus,
		Title:  &newTitle,
	}, token)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var task model.Task
	json.NewDecoder(w.Body).Decode(&task)
	if task.Status != "done" {
		t.Fatalf("expected status 'done', got %q", task.Status)
	}
	if task.Title != "Updated Task" {
		t.Fatalf("expected title 'Updated Task', got %q", task.Title)
	}
}

func TestUpdateTaskForbidden(t *testing.T) {
	r := setupTestRouter(t)
	ownerToken := registerUser(t, r, "Task Owner", "testupdatetaskforbiddenowner@test.com", "password123")
	otherToken := registerUser(t, r, "Task Other", "testupdatetaskforbiddenother@test.com", "password123")

	projectID := createProject(t, r, ownerToken, "Forbidden Task Project", "desc")
	taskID := createTask(t, r, ownerToken, projectID, "Owner Task", "todo", "medium")

	newStatus := "done"
	w := doRequest(t, r, http.MethodPatch, fmt.Sprintf("/api/tasks/%s", taskID), model.UpdateTaskRequest{
		Status: &newStatus,
	}, otherToken)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteTask(t *testing.T) {
	r := setupTestRouter(t)
	token := registerUser(t, r, "Delete Task User", "testdeletetask@test.com", "password123")
	projectID := createProject(t, r, token, "Delete Task Project", "desc")
	taskID := createTask(t, r, token, projectID, "Task to Delete", "todo", "low")

	w := doRequest(t, r, http.MethodDelete, fmt.Sprintf("/api/tasks/%s", taskID), nil, token)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteTaskAsProjectOwner(t *testing.T) {
	r := setupTestRouter(t)
	ownerToken := registerUser(t, r, "Project Owner", "testdeletetaskasowner@test.com", "password123")
	memberToken := registerUser(t, r, "Project Member", "testdeletetaskasmember@test.com", "password123")

	projectID := createProject(t, r, ownerToken, "Owner Delete Project", "desc")
	// Member creates a task in the owner's project
	taskID := createTask(t, r, memberToken, projectID, "Member Task", "todo", "medium")

	// Owner deletes the task
	w := doRequest(t, r, http.MethodDelete, fmt.Sprintf("/api/tasks/%s", taskID), nil, ownerToken)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteTaskForbidden(t *testing.T) {
	r := setupTestRouter(t)
	ownerToken := registerUser(t, r, "Task Creator", "testdeletetaskforbiddencreator@test.com", "password123")
	unrelatedToken := registerUser(t, r, "Unrelated User", "testdeletetaskforbiddenunrelated@test.com", "password123")

	projectID := createProject(t, r, ownerToken, "Forbidden Delete Project", "desc")
	taskID := createTask(t, r, ownerToken, projectID, "Protected Task", "todo", "high")

	w := doRequest(t, r, http.MethodDelete, fmt.Sprintf("/api/tasks/%s", taskID), nil, unrelatedToken)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestInvalidUUID(t *testing.T) {
	r := setupTestRouter(t)
	token := registerUser(t, r, "Invalid UUID User", "testinvaliduuid@test.com", "password123")

	w := doRequest(t, r, http.MethodGet, "/api/projects/not-a-uuid", nil, token)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestInvalidStatusFilter(t *testing.T) {
	r := setupTestRouter(t)
	token := registerUser(t, r, "Invalid Status User", "testinvalidstatus@test.com", "password123")
	projectID := createProject(t, r, token, "Invalid Status Project", "desc")

	w := doRequest(t, r, http.MethodGet, fmt.Sprintf("/api/projects/%s/tasks?status=banana", projectID), nil, token)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
