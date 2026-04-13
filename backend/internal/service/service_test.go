package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/shrinishv/taskflow-shrinish/backend/internal/model"
	"github.com/shrinishv/taskflow-shrinish/backend/internal/repository"
)

func ptr(s string) *string { return &s }

func TestAuthService(t *testing.T) {
	const jwtSecret = "test-secret"

	t.Run("Register success", func(t *testing.T) {
		userRepo := &mockUserRepo{
			CreateFn: func(ctx context.Context, name, email, hashedPassword string) (*model.User, error) {
				return &model.User{
					ID:        "user-1",
					Name:      name,
					Email:     email,
					Password:  hashedPassword,
					CreatedAt: time.Now(),
				}, nil
			},
		}

		svc := NewAuthService(userRepo, jwtSecret)
		resp, err := svc.Register(context.Background(), &model.RegisterRequest{
			Name:     "Alice",
			Email:    "alice@example.com",
			Password: "password123",
		})

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp.Token == "" {
			t.Fatal("expected non-empty token")
		}
		if resp.User.Email != "alice@example.com" {
			t.Fatalf("expected email alice@example.com, got %s", resp.User.Email)
		}
	})

	t.Run("Register duplicate email", func(t *testing.T) {
		userRepo := &mockUserRepo{
			CreateFn: func(ctx context.Context, name, email, hashedPassword string) (*model.User, error) {
				return nil, repository.ErrDuplicateEmail
			},
		}

		svc := NewAuthService(userRepo, jwtSecret)
		_, err := svc.Register(context.Background(), &model.RegisterRequest{
			Name:     "Alice",
			Email:    "alice@example.com",
			Password: "password123",
		})

		if !errors.Is(err, ErrEmailTaken) {
			t.Fatalf("expected ErrEmailTaken, got %v", err)
		}
	})

	t.Run("Login success", func(t *testing.T) {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

		userRepo := &mockUserRepo{
			FindByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
				return &model.User{
					ID:        "user-1",
					Name:      "Alice",
					Email:     email,
					Password:  string(hashed),
					CreatedAt: time.Now(),
				}, nil
			},
		}

		svc := NewAuthService(userRepo, jwtSecret)
		resp, err := svc.Login(context.Background(), &model.LoginRequest{
			Email:    "alice@example.com",
			Password: "password123",
		})

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp.Token == "" {
			t.Fatal("expected non-empty token")
		}
	})

	t.Run("Login wrong password", func(t *testing.T) {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

		userRepo := &mockUserRepo{
			FindByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
				return &model.User{
					ID:        "user-1",
					Name:      "Alice",
					Email:     email,
					Password:  string(hashed),
					CreatedAt: time.Now(),
				}, nil
			},
		}

		svc := NewAuthService(userRepo, jwtSecret)
		_, err := svc.Login(context.Background(), &model.LoginRequest{
			Email:    "alice@example.com",
			Password: "wrongpassword",
		})

		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("Login user not found", func(t *testing.T) {
		userRepo := &mockUserRepo{
			FindByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
				return nil, repository.ErrNotFound
			},
		}

		svc := NewAuthService(userRepo, jwtSecret)
		_, err := svc.Login(context.Background(), &model.LoginRequest{
			Email:    "nobody@example.com",
			Password: "password123",
		})

		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("expected ErrInvalidCredentials, got %v", err)
		}
	})
}

func TestProjectService(t *testing.T) {
	t.Run("Update by owner succeeds", func(t *testing.T) {
		projectRepo := &mockProjectRepo{
			FindByIDFn: func(ctx context.Context, id string) (*model.Project, error) {
				return &model.Project{
					ID:        "proj-1",
					Name:      "My Project",
					OwnerID:   "user-1",
					CreatedAt: time.Now(),
				}, nil
			},
			UpdateFn: func(ctx context.Context, id string, req *model.UpdateProjectRequest) (*model.Project, error) {
				return &model.Project{
					ID:        id,
					Name:      *req.Name,
					OwnerID:   "user-1",
					CreatedAt: time.Now(),
				}, nil
			},
		}
		taskRepo := &mockTaskRepo{}

		svc := NewProjectService(projectRepo, taskRepo)
		updated, err := svc.Update(context.Background(), "proj-1", "user-1", &model.UpdateProjectRequest{
			Name: ptr("Updated Name"),
		})

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if updated.Name != "Updated Name" {
			t.Fatalf("expected name 'Updated Name', got %s", updated.Name)
		}
	})

	t.Run("Update by non-owner returns ErrForbidden", func(t *testing.T) {
		projectRepo := &mockProjectRepo{
			FindByIDFn: func(ctx context.Context, id string) (*model.Project, error) {
				return &model.Project{
					ID:        "proj-1",
					Name:      "My Project",
					OwnerID:   "user-1",
					CreatedAt: time.Now(),
				}, nil
			},
		}
		taskRepo := &mockTaskRepo{}

		svc := NewProjectService(projectRepo, taskRepo)
		_, err := svc.Update(context.Background(), "proj-1", "user-2", &model.UpdateProjectRequest{
			Name: ptr("Hacked"),
		})

		if !errors.Is(err, ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("Delete by owner succeeds", func(t *testing.T) {
		projectRepo := &mockProjectRepo{
			FindByIDFn: func(ctx context.Context, id string) (*model.Project, error) {
				return &model.Project{
					ID:        "proj-1",
					Name:      "My Project",
					OwnerID:   "user-1",
					CreatedAt: time.Now(),
				}, nil
			},
			DeleteFn: func(ctx context.Context, id string) error {
				return nil
			},
		}
		taskRepo := &mockTaskRepo{}

		svc := NewProjectService(projectRepo, taskRepo)
		err := svc.Delete(context.Background(), "proj-1", "user-1")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("Delete by non-owner returns ErrForbidden", func(t *testing.T) {
		projectRepo := &mockProjectRepo{
			FindByIDFn: func(ctx context.Context, id string) (*model.Project, error) {
				return &model.Project{
					ID:        "proj-1",
					Name:      "My Project",
					OwnerID:   "user-1",
					CreatedAt: time.Now(),
				}, nil
			},
		}
		taskRepo := &mockTaskRepo{}

		svc := NewProjectService(projectRepo, taskRepo)
		err := svc.Delete(context.Background(), "proj-1", "user-2")

		if !errors.Is(err, ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})
}

func TestTaskService(t *testing.T) {
	baseTask := model.Task{
		ID:         "task-1",
		Title:      "Test Task",
		Status:     "todo",
		Priority:   "medium",
		ProjectID:  "proj-1",
		AssigneeID: ptr("user-assignee"),
		CreatedBy:  ptr("user-creator"),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	baseProject := model.Project{
		ID:        "proj-1",
		Name:      "My Project",
		OwnerID:   "user-owner",
		CreatedAt: time.Now(),
	}

	makeTaskRepo := func() *mockTaskRepo {
		return &mockTaskRepo{
			FindByIDFn: func(ctx context.Context, id string) (*model.Task, error) {
				task := baseTask
				return &task, nil
			},
			UpdateFn: func(ctx context.Context, id string, req *model.UpdateTaskRequest, assigneeID *string, assigneeChanged bool) (*model.Task, error) {
				task := baseTask
				if req.Title != nil {
					task.Title = *req.Title
				}
				return &task, nil
			},
			DeleteFn: func(ctx context.Context, id string) error {
				return nil
			},
		}
	}

	makeProjectRepo := func() *mockProjectRepo {
		return &mockProjectRepo{
			FindByIDFn: func(ctx context.Context, id string) (*model.Project, error) {
				proj := baseProject
				return &proj, nil
			},
		}
	}

	makeUserRepo := func() *mockUserRepo {
		return &mockUserRepo{
			FindByEmailFn: func(ctx context.Context, email string) (*model.User, error) {
				return &model.User{ID: "resolved-user-id", Name: "Resolved", Email: email}, nil
			},
		}
	}

	t.Run("Update by project owner succeeds", func(t *testing.T) {
		svc := NewTaskService(makeTaskRepo(), makeProjectRepo(), makeUserRepo())
		updated, err := svc.Update(context.Background(), "task-1", "user-owner", &model.UpdateTaskRequest{
			Title: ptr("New Title"),
		})

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if updated == nil {
			t.Fatal("expected updated task, got nil")
		}
	})

	t.Run("Update by task creator succeeds", func(t *testing.T) {
		svc := NewTaskService(makeTaskRepo(), makeProjectRepo(), makeUserRepo())
		updated, err := svc.Update(context.Background(), "task-1", "user-creator", &model.UpdateTaskRequest{
			Title: ptr("New Title"),
		})

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if updated == nil {
			t.Fatal("expected updated task, got nil")
		}
	})

	t.Run("Update by task assignee succeeds", func(t *testing.T) {
		svc := NewTaskService(makeTaskRepo(), makeProjectRepo(), makeUserRepo())
		updated, err := svc.Update(context.Background(), "task-1", "user-assignee", &model.UpdateTaskRequest{
			Title: ptr("New Title"),
		})

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if updated == nil {
			t.Fatal("expected updated task, got nil")
		}
	})

	t.Run("Update by unrelated user returns ErrForbidden", func(t *testing.T) {
		svc := NewTaskService(makeTaskRepo(), makeProjectRepo(), makeUserRepo())
		_, err := svc.Update(context.Background(), "task-1", "user-random", &model.UpdateTaskRequest{
			Title: ptr("New Title"),
		})

		if !errors.Is(err, ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("Delete by project owner succeeds", func(t *testing.T) {
		svc := NewTaskService(makeTaskRepo(), makeProjectRepo(), makeUserRepo())
		err := svc.Delete(context.Background(), "task-1", "user-owner")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("Delete by task creator succeeds", func(t *testing.T) {
		svc := NewTaskService(makeTaskRepo(), makeProjectRepo(), makeUserRepo())
		err := svc.Delete(context.Background(), "task-1", "user-creator")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("Delete by unrelated user returns ErrForbidden", func(t *testing.T) {
		svc := NewTaskService(makeTaskRepo(), makeProjectRepo(), makeUserRepo())
		err := svc.Delete(context.Background(), "task-1", "user-random")

		if !errors.Is(err, ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})
}
