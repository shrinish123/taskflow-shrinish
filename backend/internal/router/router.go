package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/shrinishv/taskflow-shrinish/backend/internal/handler"
	"github.com/shrinishv/taskflow-shrinish/backend/internal/middleware"
	"github.com/shrinishv/taskflow-shrinish/backend/internal/service"
)

func New(
	authService *service.AuthService,
	authHandler *handler.AuthHandler,
	projectHandler *handler.ProjectHandler,
	taskHandler *handler.TaskHandler,
	allowedOrigins []string,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(middleware.BodyLimit(1 << 20))
	r.Use(middleware.Logging)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		handler.JSON(w, 200, map[string]string{"status": "ok"})
	})

	r.Post("/api/auth/register", authHandler.Register)
	r.Post("/api/auth/login", authHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(authService))

		r.Get("/api/projects", projectHandler.List)
		r.Post("/api/projects", projectHandler.Create)
		r.Get("/api/projects/{id}", projectHandler.Get)
		r.Patch("/api/projects/{id}", projectHandler.Update)
		r.Delete("/api/projects/{id}", projectHandler.Delete)
		r.Get("/api/projects/{id}/stats", projectHandler.Stats)

		r.Get("/api/projects/{id}/tasks", taskHandler.List)
		r.Post("/api/projects/{id}/tasks", taskHandler.Create)
		r.Patch("/api/tasks/{id}", taskHandler.Update)
		r.Delete("/api/tasks/{id}", taskHandler.Delete)
	})

	return r
}
