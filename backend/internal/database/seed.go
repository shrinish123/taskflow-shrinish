package database

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// Seed inserts test data (idempotent — skips if seed user exists).
func Seed(ctx context.Context, pool *pgxpool.Pool) error {
	var exists bool
	err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", "test@example.com").Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		slog.Info("seed data already exists, skipping")
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), 12)
	if err != nil {
		return err
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	userID := uuid.Must(uuid.NewV7()).String()
	_, err = tx.Exec(ctx,
		`INSERT INTO users (id, name, email, password) VALUES ($1, $2, $3, $4)`,
		userID, "Test User", "test@example.com", string(hash),
	)
	if err != nil {
		return err
	}

	projectID := uuid.Must(uuid.NewV7()).String()
	_, err = tx.Exec(ctx,
		`INSERT INTO projects (id, name, description, owner_id) VALUES ($1, $2, $3, $4)`,
		projectID, "My First Project", "A sample project to get started with TaskFlow", userID,
	)
	if err != nil {
		return err
	}

	tasks := []struct {
		title       string
		description string
		status      string
		priority    string
	}{
		{"Set up development environment", "Install Go, Node.js, and Docker", "done", "high"},
		{"Design the database schema", "Create ERD and write migration files", "in_progress", "medium"},
		{"Write API documentation", "Document all endpoints with examples", "todo", "low"},
	}

	for _, t := range tasks {
		taskID := uuid.Must(uuid.NewV7()).String()
		_, err = tx.Exec(ctx,
			`INSERT INTO tasks (id, title, description, status, priority, project_id, assignee_id, created_by)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			taskID, t.title, t.description, t.status, t.priority, projectID, userID, userID,
		)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	slog.Info("seed data inserted successfully")
	return nil
}
