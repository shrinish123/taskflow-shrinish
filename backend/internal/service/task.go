package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/shrinishv/taskflow-shrinish/backend/internal/model"
	"github.com/shrinishv/taskflow-shrinish/backend/internal/repository"
)

type TaskService struct {
	taskRepo    TaskRepository
	projectRepo ProjectRepository
	userRepo    UserRepository
}

func NewTaskService(taskRepo TaskRepository, projectRepo ProjectRepository, userRepo UserRepository) *TaskService {
	return &TaskService{
		taskRepo:    taskRepo,
		projectRepo: projectRepo,
		userRepo:    userRepo,
	}
}

func (s *TaskService) resolveAssigneeEmail(ctx context.Context, email *string) (*string, error) {
	if email == nil {
		return nil, nil
	}
	user, err := s.userRepo.FindByEmail(ctx, *email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, &AppError{
				Status:  http.StatusBadRequest,
				Message: "assignee not found: no user with email " + *email,
				Err:     err,
			}
		}
		return nil, InternalErr(err)
	}
	return &user.ID, nil
}

func (s *TaskService) List(ctx context.Context, projectID string, filter model.TaskFilter) ([]model.Task, int, error) {
	if _, err := s.projectRepo.FindByID(ctx, projectID); err != nil {
		return nil, 0, mapRepoErr(err, "project")
	}
	tasks, total, err := s.taskRepo.ListByProject(ctx, projectID, filter)
	if err != nil {
		return nil, 0, mapRepoErr(err, "task")
	}
	return tasks, total, nil
}

func (s *TaskService) Create(ctx context.Context, projectID string, req *model.CreateTaskRequest, userID string) (*model.Task, error) {
	if _, err := s.projectRepo.FindByID(ctx, projectID); err != nil {
		return nil, mapRepoErr(err, "project")
	}

	assigneeID, err := s.resolveAssigneeEmail(ctx, req.AssigneeEmail)
	if err != nil {
		return nil, err
	}

	task, err := s.taskRepo.Create(ctx, projectID, req, userID, assigneeID)
	if err != nil {
		return nil, mapRepoErr(err, "task")
	}
	return task, nil
}

func (s *TaskService) Update(ctx context.Context, taskID, userID string, req *model.UpdateTaskRequest) (*model.Task, error) {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, mapRepoErr(err, "task")
	}

	project, err := s.projectRepo.FindByID(ctx, task.ProjectID)
	if err != nil {
		return nil, mapRepoErr(err, "project")
	}

	isOwner := project.OwnerID == userID
	isCreator := task.CreatedBy != nil && *task.CreatedBy == userID
	isAssignee := task.AssigneeID != nil && *task.AssigneeID == userID
	if !isOwner && !isCreator && !isAssignee {
		return nil, ForbiddenErr("you do not have permission to update this task")
	}

	assigneeChanged := req.AssigneeEmail != nil
	var assigneeID *string
	if assigneeChanged {
		if *req.AssigneeEmail == "" {
			assigneeID = nil // unassign
		} else {
			resolved, err := s.resolveAssigneeEmail(ctx, req.AssigneeEmail)
			if err != nil {
				return nil, err
			}
			assigneeID = resolved
		}
	}

	result, err := s.taskRepo.Update(ctx, taskID, req, assigneeID, assigneeChanged)
	if err != nil {
		return nil, mapRepoErr(err, "task")
	}
	return result, nil
}

func (s *TaskService) Delete(ctx context.Context, taskID, userID string) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return mapRepoErr(err, "task")
	}

	project, err := s.projectRepo.FindByID(ctx, task.ProjectID)
	if err != nil {
		return mapRepoErr(err, "project")
	}

	isOwner := project.OwnerID == userID
	isCreator := task.CreatedBy != nil && *task.CreatedBy == userID
	if !isOwner && !isCreator {
		return ForbiddenErr("you do not have permission to delete this task")
	}

	if err := s.taskRepo.Delete(ctx, taskID); err != nil {
		return mapRepoErr(err, "task")
	}
	return nil
}
