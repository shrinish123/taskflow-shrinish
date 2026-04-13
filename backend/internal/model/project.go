package model

import "time"

type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type ProjectWithTasks struct {
	Project
	Tasks []Task `json:"tasks"`
}

type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (r *CreateProjectRequest) Validate() map[string]string {
	errs := make(map[string]string)
	if r.Name == "" {
		errs["name"] = "is required"
	}
	return errs
}

type UpdateProjectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func (r *UpdateProjectRequest) Validate() map[string]string {
	errs := make(map[string]string)
	if r.Name != nil && *r.Name == "" {
		errs["name"] = "cannot be empty"
	}
	return errs
}

type ProjectStats struct {
	ProjectID    string              `json:"project_id"`
	TotalTasks   int                 `json:"total_tasks"`
	ByStatus     map[string]int      `json:"by_status"`
	ByAssignee   []AssigneeTaskCount `json:"by_assignee"`
}

type AssigneeTaskCount struct {
	AssigneeID   *string `json:"assignee_id"`
	AssigneeName string  `json:"assignee_name"`
	Count        int     `json:"count"`
}

type ProjectFilter struct {
	Page  int
	Limit int
}
