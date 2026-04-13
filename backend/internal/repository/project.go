package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shrinishv/taskflow-shrinish/backend/internal/model"
)

type ProjectRepository struct {
	pool *pgxpool.Pool
}

func NewProjectRepository(pool *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{pool: pool}
}

func (r *ProjectRepository) ListByUser(ctx context.Context, userID string, filter model.ProjectFilter) ([]model.Project, int, error) {
	baseQuery := `SELECT id, name, description, owner_id, created_at FROM projects WHERE owner_id = $1
		 UNION
		 SELECT p.id, p.name, p.description, p.owner_id, p.created_at
		 FROM projects p
		 WHERE EXISTS (SELECT 1 FROM tasks t WHERE t.project_id = p.id AND t.assignee_id = $1)`

	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM ("+baseQuery+") sub", userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := baseQuery + " ORDER BY created_at DESC"
	args := []any{userID}
	argIdx := 2

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filter.Limit)
		argIdx++

		if filter.Page > 1 {
			offset := (filter.Page - 1) * filter.Limit
			query += fmt.Sprintf(" OFFSET $%d", argIdx)
			args = append(args, offset)
		}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var projects []model.Project
	for rows.Next() {
		var p model.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt); err != nil {
			return nil, 0, err
		}
		projects = append(projects, p)
	}
	return projects, total, rows.Err()
}

func (r *ProjectRepository) Create(ctx context.Context, name, description, ownerID string) (*model.Project, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	p := &model.Project{}
	err = r.pool.QueryRow(ctx,
		`INSERT INTO projects (id, name, description, owner_id)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, name, description, owner_id, created_at`,
		id.String(), name, description, ownerID,
	).Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *ProjectRepository) FindByID(ctx context.Context, id string) (*model.Project, error) {
	p := &model.Project{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, description, owner_id, created_at FROM projects WHERE id = $1`,
		id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *ProjectRepository) Update(ctx context.Context, id string, req *model.UpdateProjectRequest) (*model.Project, error) {
	p := &model.Project{}
	err := r.pool.QueryRow(ctx,
		`UPDATE projects
		 SET name        = CASE WHEN $1::boolean THEN $2 ELSE name END,
		     description = CASE WHEN $3::boolean THEN $4 ELSE description END
		 WHERE id = $5
		 RETURNING id, name, description, owner_id, created_at`,
		req.Name != nil, req.Name,
		req.Description != nil, req.Description,
		id,
	).Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return p, nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
