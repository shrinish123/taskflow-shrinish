package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/shrinishv/taskflow-shrinish/backend/internal/middleware"
	"github.com/shrinishv/taskflow-shrinish/backend/internal/model"
	"github.com/shrinishv/taskflow-shrinish/backend/internal/service"
)

type TaskHandler struct {
	taskService *service.TaskService
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID, ok := ValidID(w, r, "id")
	if !ok {
		return
	}

	status := r.URL.Query().Get("status")
	assignee := r.URL.Query().Get("assignee")

	filterErrs := make(map[string]string)
	if status != "" && status != "todo" && status != "in_progress" && status != "done" {
		filterErrs["status"] = "must be one of: todo, in_progress, done"
	}
	if assignee != "" && !IsValidUUID(assignee) {
		filterErrs["assignee"] = "must be a valid UUID"
	}
	if len(filterErrs) > 0 {
		ValidationError(w, filterErrs)
		return
	}

	filter := model.TaskFilter{
		Status:     status,
		AssigneeID: assignee,
	}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			filter.Page = p
		}
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			filter.Limit = l
		}
	}
	if filter.Page > 0 && filter.Limit == 0 {
		filter.Limit = 20
	}

	tasks, total, err := h.taskService.List(r.Context(), projectID, filter)
	if err != nil {
		WriteError(w, err)
		return
	}

	if tasks == nil {
		tasks = []model.Task{}
	}

	JSON(w, http.StatusOK, map[string]any{
		"tasks": tasks,
		"total": total,
		"page":  filter.Page,
		"limit": filter.Limit,
	})
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	projectID, ok := ValidID(w, r, "id")
	if !ok {
		return
	}

	var req model.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		ValidationError(w, errs)
		return
	}

	task, err := h.taskService.Create(r.Context(), projectID, &req, userID)
	if err != nil {
		WriteError(w, err)
		return
	}

	JSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	taskID, ok := ValidID(w, r, "id")
	if !ok {
		return
	}

	var req model.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		ValidationError(w, errs)
		return
	}

	task, err := h.taskService.Update(r.Context(), taskID, userID, &req)
	if err != nil {
		WriteError(w, err)
		return
	}

	JSON(w, http.StatusOK, task)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	taskID, ok := ValidID(w, r, "id")
	if !ok {
		return
	}

	if err := h.taskService.Delete(r.Context(), taskID, userID); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
