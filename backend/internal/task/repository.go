package task

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-sql-driver/mysql"
)

// ListFilter is the repository-level view of the list query.
// Sort arrives already normalized ("field:asc|desc") from the service.
type ListFilter struct {
	Status   string
	Keyword  string
	Assignee string
	Page     int
	Limit    int
	Sort     string
}

// Repository persists and retrieves tasks.
type Repository interface {
	List(ctx context.Context, f ListFilter) (items []Task, total int, err error)
	Create(ctx context.Context, t *Task) (*Task, error)
	Update(ctx context.Context, t *Task) (*Task, error)
	SoftDelete(ctx context.Context, id int64) error
}

type mysqlRepository struct {
	db *sql.DB
}

// NewRepository returns a MySQL-backed Repository.
func NewRepository(db *sql.DB) Repository {
	return &mysqlRepository{db: db}
}

const taskColumns = "id, title, description, status, assignee, due_date, created_at, updated_at"

// sortableColumns whitelists every column ORDER BY may use. ORDER BY cannot be
// parameterized, so any value outside this map falls back to the default.
var sortableColumns = map[string]string{
	"title":      "title",
	"status":     "status",
	"assignee":   "assignee",
	"created_at": "created_at",
	"updated_at": "updated_at",
}

func orderClause(sort string) string {
	field, dir := "created_at", "DESC"
	if parts := strings.SplitN(sort, ":", 2); len(parts) > 0 {
		if col, ok := sortableColumns[parts[0]]; ok {
			field = col
		}
		if len(parts) == 2 && strings.EqualFold(parts[1], "asc") {
			dir = "ASC"
		}
	}
	return field + " " + dir
}

func (r *mysqlRepository) List(ctx context.Context, f ListFilter) ([]Task, int, error) {
	// Soft-deleted tasks are excluded from every read (Task 4).
	where := " WHERE deleted_at IS NULL"
	var args []any
	if f.Status != "" {
		where += " AND status = ?"
		args = append(args, f.Status)
	}
	if f.Keyword != "" {
		where += " AND title LIKE CONCAT('%', ?, '%')"
		args = append(args, f.Keyword)
	}
	if f.Assignee != "" {
		where += " AND assignee = ?"
		args = append(args, f.Assignee)
	}

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tasks"+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count tasks: %w", err)
	}

	offset := (f.Page - 1) * f.Limit
	query := "SELECT " + taskColumns + " FROM tasks" + where +
		" ORDER BY " + orderClause(f.Sort) + " LIMIT ? OFFSET ?"
	rows, err := r.db.QueryContext(ctx, query, append(args, f.Limit, offset)...)
	if err != nil {
		return nil, 0, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	items := make([]Task, 0)
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Assignee, &t.DueDate, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan task: %w", err)
		}
		items = append(items, t)
	}
	return items, total, rows.Err()
}

// Create inserts the task and re-selects it so the response carries the real
// generated id and timestamps from the database.
func (r *mysqlRepository) Create(ctx context.Context, t *Task) (*Task, error) {
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO tasks (title, description, status, assignee, due_date) VALUES (?, ?, ?, ?, ?)",
		t.Title, t.Description, t.Status, t.Assignee, t.DueDate)
	if err != nil {
		if isDuplicate(err) {
			return nil, ErrDuplicateTitle
		}
		return nil, fmt.Errorf("create task: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("create task last insert id: %w", err)
	}
	t.ID = id
	return r.getByID(ctx, t.ID)
}

func (r *mysqlRepository) Update(ctx context.Context, t *Task) (*Task, error) {
	res, err := r.db.ExecContext(ctx,
		"UPDATE tasks SET title = ?, description = ?, status = ?, assignee = ?, due_date = ?"+
			" WHERE id = ? AND deleted_at IS NULL",
		t.Title, t.Description, t.Status, t.Assignee, t.DueDate, t.ID)
	if err != nil {
		if isDuplicate(err) {
			return nil, ErrDuplicateTitle
		}
		return nil, fmt.Errorf("update task: %w", err)
	}
	if n, err := res.RowsAffected(); err != nil {
		return nil, fmt.Errorf("update task rows affected: %w", err)
	} else if n == 0 {
		return nil, ErrNotFound
	}
	return r.getByID(ctx, t.ID)
}

func (r *mysqlRepository) SoftDelete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx,
		"UPDATE tasks SET deleted_at = NOW() WHERE id = ? AND deleted_at IS NULL", id)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	if n, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("delete task rows affected: %w", err)
	} else if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *mysqlRepository) getByID(ctx context.Context, id int64) (*Task, error) {
	var t Task
	err := r.db.QueryRowContext(ctx,
		"SELECT "+taskColumns+" FROM tasks WHERE id = ? AND deleted_at IS NULL", id).
		Scan(&t.ID, &t.Title, &t.Description, &t.Status, &t.Assignee, &t.DueDate, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	return &t, nil
}

// isDuplicate reports whether err is MySQL error 1062 (duplicate key),
// so a UNIQUE(title) violation becomes ErrDuplicateTitle instead of a 500.
func isDuplicate(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
