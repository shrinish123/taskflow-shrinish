package service

import (
	"context"

	"github.com/shrinishv/taskflow-shrinish/backend/internal/model"
)

type ProjectService struct {
	projectRepo ProjectRepository
	taskRepo    TaskRepository
}

func NewProjectService(projectRepo ProjectRepository, taskRepo TaskRepository) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		taskRepo:    taskRepo,
	}
}

func (s *ProjectService) List(ctx context.Context, userID string, filter model.ProjectFilter) ([]model.Project, int, error) {
	projects, total, err := s.projectRepo.ListByUser(ctx, userID, filter)
	if err != nil {
		return nil, 0, mapRepoErr(err, "project")
	}
	return projects, total, nil
}

func (s *ProjectService) Create(ctx context.Context, req *model.CreateProjectRequest, ownerID string) (*model.Project, error) {
	project, err := s.projectRepo.Create(ctx, req.Name, req.Description, ownerID)
	if err != nil {
		return nil, mapRepoErr(err, "project")
	}
	return project, nil
}

func (s *ProjectService) GetWithTasks(ctx context.Context, projectID string) (*model.ProjectWithTasks, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return nil, mapRepoErr(err, "project")
	}

	tasks, _, err := s.taskRepo.ListByProject(ctx, projectID, model.TaskFilter{})
	if err != nil {
		return nil, mapRepoErr(err, "task")
	}
	if tasks == nil {
		tasks = []model.Task{}
	}

	return &model.ProjectWithTasks{
		Project: *project,
		Tasks:   tasks,
	}, nil
}

func (s *ProjectService) Update(ctx context.Context, projectID, userID string, req *model.UpdateProjectRequest) (*model.Project, error) {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return nil, mapRepoErr(err, "project")
	}

	if project.OwnerID != userID {
		return nil, ForbiddenErr("only the project owner can update this project")
	}

	result, err := s.projectRepo.Update(ctx, projectID, req)
	if err != nil {
		return nil, mapRepoErr(err, "project")
	}
	return result, nil
}

func (s *ProjectService) Delete(ctx context.Context, projectID, userID string) error {
	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return mapRepoErr(err, "project")
	}

	if project.OwnerID != userID {
		return ForbiddenErr("only the project owner can delete this project")
	}

	if err := s.projectRepo.Delete(ctx, projectID); err != nil {
		return mapRepoErr(err, "project")
	}
	return nil
}

func (s *ProjectService) GetStats(ctx context.Context, projectID string) (*model.ProjectStats, error) {
	if _, err := s.projectRepo.FindByID(ctx, projectID); err != nil {
		return nil, mapRepoErr(err, "project")
	}
	stats, err := s.taskRepo.StatsByProject(ctx, projectID)
	if err != nil {
		return nil, mapRepoErr(err, "project")
	}
	return stats, nil
}
