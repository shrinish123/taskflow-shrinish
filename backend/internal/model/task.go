package model

import (
	"net/mail"
	"time"
)

type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	ProjectID   string    `json:"project_id"`
	AssigneeID  *string   `json:"assignee_id"`
	CreatedBy   *string   `json:"created_by"`
	DueDate     *string   `json:"due_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

var validStatuses = map[string]bool{"todo": true, "in_progress": true, "done": true}
var validPriorities = map[string]bool{"low": true, "medium": true, "high": true}

func isValidEmail(s string) bool {
	_, err := mail.ParseAddress(s)
	return err == nil
}

func isValidDate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

type CreateTaskRequest struct {
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	Status        string  `json:"status"`
	Priority      string  `json:"priority"`
	AssigneeEmail *string `json:"assignee_email"`
	DueDate       *string `json:"due_date"`
}

func (r *CreateTaskRequest) Validate() map[string]string {
	errs := make(map[string]string)

	if r.Title == "" {
		errs["title"] = "is required"
	}

	if r.Status == "" {
		r.Status = "todo"
	} else if !validStatuses[r.Status] {
		errs["status"] = "must be one of: todo, in_progress, done"
	}

	if r.Priority == "" {
		r.Priority = "medium"
	} else if !validPriorities[r.Priority] {
		errs["priority"] = "must be one of: low, medium, high"
	}

	if r.AssigneeEmail != nil && *r.AssigneeEmail == "" {
		r.AssigneeEmail = nil
	}
	if r.DueDate != nil && *r.DueDate == "" {
		r.DueDate = nil
	}

	if r.AssigneeEmail != nil && !isValidEmail(*r.AssigneeEmail) {
		errs["assignee_email"] = "must be a valid email address"
	}
	if r.DueDate != nil && !isValidDate(*r.DueDate) {
		errs["due_date"] = "must be a valid date (YYYY-MM-DD)"
	}

	return errs
}

type UpdateTaskRequest struct {
	Title         *string `json:"title"`
	Description   *string `json:"description"`
	Status        *string `json:"status"`
	Priority      *string `json:"priority"`
	AssigneeEmail *string `json:"assignee_email"`
	DueDate       *string `json:"due_date"`
}

func (r *UpdateTaskRequest) Validate() map[string]string {
	errs := make(map[string]string)

	if r.Title != nil && *r.Title == "" {
		errs["title"] = "cannot be empty"
	}
	if r.Status != nil && !validStatuses[*r.Status] {
		errs["status"] = "must be one of: todo, in_progress, done"
	}
	if r.Priority != nil && !validPriorities[*r.Priority] {
		errs["priority"] = "must be one of: low, medium, high"
	}
	if r.AssigneeEmail != nil && *r.AssigneeEmail == "" {
		// empty = unassign, skip validation
	} else if r.AssigneeEmail != nil && !isValidEmail(*r.AssigneeEmail) {
		errs["assignee_email"] = "must be a valid email address"
	}

	if r.DueDate != nil && *r.DueDate == "" {
		// empty = clear due date, skip validation
	} else if r.DueDate != nil && !isValidDate(*r.DueDate) {
		errs["due_date"] = "must be a valid date (YYYY-MM-DD)"
	}

	return errs
}

type TaskFilter struct {
	Status     string
	AssigneeID string
	Page       int
	Limit      int
}
