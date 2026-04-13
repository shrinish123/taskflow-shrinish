package service

import (
	"context"

	"github.com/shrinishv/taskflow-shrinish/backend/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, name, email, hashedPassword string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
}

type ProjectRepository interface {
	ListByUser(ctx context.Context, userID string, filter model.ProjectFilter) ([]model.Project, int, error)
	Create(ctx context.Context, name, description, ownerID string) (*model.Project, error)
	FindByID(ctx context.Context, id string) (*model.Project, error)
	Update(ctx context.Context, id string, req *model.UpdateProjectRequest) (*model.Project, error)
	Delete(ctx context.Context, id string) error
}

type TaskRepository interface {
	ListByProject(ctx context.Context, projectID string, filter model.TaskFilter) ([]model.Task, int, error)
	Create(ctx context.Context, projectID string, req *model.CreateTaskRequest, creatorID string, assigneeID *string) (*model.Task, error)
	FindByID(ctx context.Context, id string) (*model.Task, error)
	Update(ctx context.Context, id string, req *model.UpdateTaskRequest, assigneeID *string, assigneeChanged bool) (*model.Task, error)
	Delete(ctx context.Context, id string) error
	StatsByProject(ctx context.Context, projectID string) (*model.ProjectStats, error)
}
