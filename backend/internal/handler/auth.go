package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/shrinishv/taskflow-shrinish/backend/internal/model"
	"github.com/shrinishv/taskflow-shrinish/backend/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		ValidationError(w, errs)
		return
	}

	resp, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			ValidationError(w, map[string]string{"email": "is already taken"})
			return
		}
		WriteError(w, err)
		return
	}

	JSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if errs := req.Validate(); len(errs) > 0 {
		ValidationError(w, errs)
		return
	}

	resp, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		WriteError(w, err)
		return
	}

	JSON(w, http.StatusOK, resp)
}
