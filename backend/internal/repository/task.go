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

type TaskRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{pool: pool}
}

func (r *TaskRepository) ListByProject(ctx context.Context, projectID string, filter model.TaskFilter) ([]model.Task, int, error) {
	query := `SELECT id, title, description, status, priority, project_id, assignee_id,
	                 created_by, due_date::text, created_at, updated_at
	          FROM tasks WHERE project_id = $1`
	countQuery := `SELECT COUNT(*) FROM tasks WHERE project_id = $1`
	args := []any{projectID}
	argIdx := 2

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		countQuery += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.AssigneeID != "" {
		query += fmt.Sprintf(" AND assignee_id = $%d", argIdx)
		countQuery += fmt.Sprintf(" AND assignee_id = $%d", argIdx)
		args = append(args, filter.AssigneeID)
		argIdx++
	}

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query += " ORDER BY created_at DESC"

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

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
			&t.ProjectID, &t.AssigneeID, &t.CreatedBy, &t.DueDate, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, t)
	}
	return tasks, total, rows.Err()
}

func (r *TaskRepository) Create(ctx context.Context, projectID string, req *model.CreateTaskRequest, creatorID string, assigneeID *string) (*model.Task, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	t := &model.Task{}
	err = r.pool.QueryRow(ctx,
		`INSERT INTO tasks (id, title, description, status, priority, project_id, assignee_id, created_by, due_date)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::date)
		 RETURNING id, title, description, status, priority, project_id, assignee_id,
		           created_by, due_date::text, created_at, updated_at`,
		id.String(), req.Title, req.Description, req.Status, req.Priority, projectID, assigneeID, creatorID, req.DueDate,
	).Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.ProjectID, &t.AssigneeID, &t.CreatedBy, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TaskRepository) FindByID(ctx context.Context, id string) (*model.Task, error) {
	t := &model.Task{}
	err := r.pool.QueryRow(ctx,
		`SELECT id, title, description, status, priority, project_id, assignee_id,
		        created_by, due_date::text, created_at, updated_at
		 FROM tasks WHERE id = $1`,
		id,
	).Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.ProjectID, &t.AssigneeID, &t.CreatedBy, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return t, nil
}

func (r *TaskRepository) Update(ctx context.Context, id string, req *model.UpdateTaskRequest, assigneeID *string, assigneeChanged bool) (*model.Task, error) {
	t := &model.Task{}
	err := r.pool.QueryRow(ctx,
		`UPDATE tasks
		 SET title       = COALESCE($1, title),
		     description = COALESCE($2, description),
		     status      = COALESCE($3, status),
		     priority    = COALESCE($4, priority),
		     assignee_id = CASE WHEN $5::boolean THEN $6::uuid ELSE assignee_id END,
		     due_date    = CASE WHEN $7::boolean THEN $8::date ELSE due_date END,
		     updated_at  = NOW()
		 WHERE id = $9
		 RETURNING id, title, description, status, priority, project_id, assignee_id,
		           created_by, due_date::text, created_at, updated_at`,
		req.Title, req.Description, req.Status, req.Priority,
		assigneeChanged, assigneeID,
		req.DueDate != nil, req.DueDate,
		id,
	).Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Priority,
		&t.ProjectID, &t.AssigneeID, &t.CreatedBy, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return t, nil
}

func (r *TaskRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *TaskRepository) StatsByProject(ctx context.Context, projectID string) (*model.ProjectStats, error) {
	stats := &model.ProjectStats{
		ProjectID: projectID,
		ByStatus:  make(map[string]int),
	}

	rows, err := r.pool.Query(ctx,
		`SELECT status, COUNT(*) FROM tasks WHERE project_id = $1 GROUP BY status`,
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats.ByStatus[status] = count
		stats.TotalTasks += count
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	rows2, err := r.pool.Query(ctx,
		`SELECT t.assignee_id, COALESCE(u.name, 'Unassigned'), COUNT(*)
		 FROM tasks t
		 LEFT JOIN users u ON u.id = t.assignee_id
		 WHERE t.project_id = $1
		 GROUP BY t.assignee_id, u.name`,
		projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()

	for rows2.Next() {
		var ac model.AssigneeTaskCount
		if err := rows2.Scan(&ac.AssigneeID, &ac.AssigneeName, &ac.Count); err != nil {
			return nil, err
		}
		stats.ByAssignee = append(stats.ByAssignee, ac)
	}

	return stats, rows2.Err()
}
