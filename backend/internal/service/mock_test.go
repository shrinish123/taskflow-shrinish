package service

import (
	"context"

	"github.com/shrinishv/taskflow-shrinish/backend/internal/model"
)

// mockUserRepo implements UserRepository with function fields for per-test customization.
type mockUserRepo struct {
	CreateFn      func(ctx context.Context, name, email, hashedPassword string) (*model.User, error)
	FindByEmailFn func(ctx context.Context, email string) (*model.User, error)
	FindByIDFn    func(ctx context.Context, id string) (*model.User, error)
}

func (m *mockUserRepo) Create(ctx context.Context, name, email, hashedPassword string) (*model.User, error) {
	return m.CreateFn(ctx, name, email, hashedPassword)
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	return m.FindByEmailFn(ctx, email)
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*model.User, error) {
	return m.FindByIDFn(ctx, id)
}

// mockProjectRepo implements ProjectRepository with function fields for per-test customization.
type mockProjectRepo struct {
	ListByUserFn func(ctx context.Context, userID string, filter model.ProjectFilter) ([]model.Project, int, error)
	CreateFn     func(ctx context.Context, name, description, ownerID string) (*model.Project, error)
	FindByIDFn   func(ctx context.Context, id string) (*model.Project, error)
	UpdateFn     func(ctx context.Context, id string, req *model.UpdateProjectRequest) (*model.Project, error)
	DeleteFn     func(ctx context.Context, id string) error
}

func (m *mockProjectRepo) ListByUser(ctx context.Context, userID string, filter model.ProjectFilter) ([]model.Project, int, error) {
	return m.ListByUserFn(ctx, userID, filter)
}

func (m *mockProjectRepo) Create(ctx context.Context, name, description, ownerID string) (*model.Project, error) {
	return m.CreateFn(ctx, name, description, ownerID)
}

func (m *mockProjectRepo) FindByID(ctx context.Context, id string) (*model.Project, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *mockProjectRepo) Update(ctx context.Context, id string, req *model.UpdateProjectRequest) (*model.Project, error) {
	return m.UpdateFn(ctx, id, req)
}

func (m *mockProjectRepo) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

// mockTaskRepo implements TaskRepository with function fields for per-test customization.
type mockTaskRepo struct {
	ListByProjectFn  func(ctx context.Context, projectID string, filter model.TaskFilter) ([]model.Task, int, error)
	CreateFn         func(ctx context.Context, projectID string, req *model.CreateTaskRequest, creatorID string, assigneeID *string) (*model.Task, error)
	FindByIDFn       func(ctx context.Context, id string) (*model.Task, error)
	UpdateFn         func(ctx context.Context, id string, req *model.UpdateTaskRequest, assigneeID *string, assigneeChanged bool) (*model.Task, error)
	DeleteFn         func(ctx context.Context, id string) error
	StatsByProjectFn func(ctx context.Context, projectID string) (*model.ProjectStats, error)
}

func (m *mockTaskRepo) ListByProject(ctx context.Context, projectID string, filter model.TaskFilter) ([]model.Task, int, error) {
	return m.ListByProjectFn(ctx, projectID, filter)
}

func (m *mockTaskRepo) Create(ctx context.Context, projectID string, req *model.CreateTaskRequest, creatorID string, assigneeID *string) (*model.Task, error) {
	return m.CreateFn(ctx, projectID, req, creatorID, assigneeID)
}

func (m *mockTaskRepo) FindByID(ctx context.Context, id string) (*model.Task, error) {
	return m.FindByIDFn(ctx, id)
}

func (m *mockTaskRepo) Update(ctx context.Context, id string, req *model.UpdateTaskRequest, assigneeID *string, assigneeChanged bool) (*model.Task, error) {
	return m.UpdateFn(ctx, id, req, assigneeID, assigneeChanged)
}

func (m *mockTaskRepo) Delete(ctx context.Context, id string) error {
	return m.DeleteFn(ctx, id)
}

func (m *mockTaskRepo) StatsByProject(ctx context.Context, projectID string) (*model.ProjectStats, error) {
	return m.StatsByProjectFn(ctx, projectID)
}
