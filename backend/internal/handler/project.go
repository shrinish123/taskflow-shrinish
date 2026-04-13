package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/shrinishv/taskflow-shrinish/backend/internal/middleware"
	"github.com/shrinishv/taskflow-shrinish/backend/internal/model"
	"github.com/shrinishv/taskflow-shrinish/backend/internal/service"
)

type ProjectHandler struct {
	projectService *service.ProjectService
}

func NewProjectHandler(projectService *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: projectService}
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var filter model.ProjectFilter
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

	projects, total, err := h.projectService.List(r.Context(), userID, filter)
	if err != nil {
		WriteError(w, err)
		return
	}

	if projects == nil {
		projects = []model.Project{}
	}

	JSON(w, http.StatusOK, map[string]any{
		"projects": projects,
		"total":    total,
		"page":     filter.Page,
		"limit":    filter.Limit,
	})
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())

	var req model.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		ValidationError(w, errs)
		return
	}

	project, err := h.projectService.Create(r.Context(), &req, userID)
	if err != nil {
		WriteError(w, err)
		return
	}

	JSON(w, http.StatusCreated, project)
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	projectID, ok := ValidID(w, r, "id")
	if !ok {
		return
	}

	project, err := h.projectService.GetWithTasks(r.Context(), projectID)
	if err != nil {
		WriteError(w, err)
		return
	}

	JSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	projectID, ok := ValidID(w, r, "id")
	if !ok {
		return
	}

	var req model.UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		ValidationError(w, errs)
		return
	}

	project, err := h.projectService.Update(r.Context(), projectID, userID, &req)
	if err != nil {
		WriteError(w, err)
		return
	}

	JSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	projectID, ok := ValidID(w, r, "id")
	if !ok {
		return
	}

	if err := h.projectService.Delete(r.Context(), projectID, userID); err != nil {
		WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ProjectHandler) Stats(w http.ResponseWriter, r *http.Request) {
	projectID, ok := ValidID(w, r, "id")
	if !ok {
		return
	}

	stats, err := h.projectService.GetStats(r.Context(), projectID)
	if err != nil {
		WriteError(w, err)
		return
	}

	JSON(w, http.StatusOK, stats)
}
