package task

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
)

func newRepoWithMock(t *testing.T) (Repository, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewRepository(db), mock
}

func oneTaskRow() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "title", "description", "status", "assignee", "due_date", "created_at", "updated_at"}).
		AddRow(1, "first", "desc", StatusTodo, "budi",
			time.Date(2026, 9, 27, 0, 0, 0, 0, time.UTC),
			time.Date(2026, 9, 27, 10, 0, 0, 0, time.UTC),
			time.Date(2026, 9, 27, 10, 0, 0, 0, time.UTC))
}

// Task 5: search — every enabled filter must appear in the WHERE clause with
// its arguments in order, pagination must page correctly, and the row set is
// always restricted to non-deleted tasks (deleted_at IS NULL).
func TestListAppliesAllFilters(t *testing.T) {
	repo, mock := newRepoWithMock(t)
	ctx := context.Background()
	f := ListFilter{Status: StatusTodo, Keyword: "fix", Assignee: "budi", Page: 2, Limit: 5, Sort: "created_at:desc"}

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM tasks WHERE deleted_at IS NULL AND status = \? AND title LIKE CONCAT\('%', \?, '%'\) AND assignee = \?$`).
		WithArgs(StatusTodo, "fix", "budi").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(42))
	mock.ExpectQuery(`SELECT id, title, description, status, assignee, due_date, created_at, updated_at FROM tasks WHERE deleted_at IS NULL AND status = \? AND title LIKE CONCAT\('%', \?, '%'\) AND assignee = \? ORDER BY created_at DESC LIMIT \? OFFSET \?$`).
		WithArgs(StatusTodo, "fix", "budi", 5, 5).
		WillReturnRows(oneTaskRow())

	items, total, err := repo.List(ctx, f)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 42 {
		t.Fatalf("total = %d, want 42", total)
	}
	if len(items) != 1 || items[0].ID != 1 || items[0].Title != "first" {
		t.Fatalf("items = %+v, want the single scanned task", items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// No filters: only the soft-delete guard plus LIMIT/OFFSET remain, and an
// empty result set is returned as an empty slice (not nil).
func TestListWithoutFiltersExcludesSoftDeleted(t *testing.T) {
	repo, mock := newRepoWithMock(t)
	f := ListFilter{Page: 1, Limit: 10, Sort: "created_at:desc"}

	mock.ExpectQuery(`^SELECT COUNT\(\*\) FROM tasks WHERE deleted_at IS NULL$`).
		WithArgs().
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(`^SELECT id, title, description, status, assignee, due_date, created_at, updated_at FROM tasks WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT \? OFFSET \?$`).
		WithArgs(10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "status", "assignee", "due_date", "created_at", "updated_at"}))

	items, total, err := repo.List(context.Background(), f)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 0 || len(items) != 0 {
		t.Fatalf("total = %d, items = %d, want 0/0", total, len(items))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// ORDER BY cannot be parameterized: whitelisted fields are used as-is,
// anything else (including injection attempts) falls back to the default.
func TestListSortWhitelist(t *testing.T) {
	tests := []struct {
		name     string
		sort     string
		expected string
	}{
		{"whitelisted asc", "title:asc", "ORDER BY title ASC"},
		{"whitelisted default dir", "updated_at", "ORDER BY updated_at DESC"},
		{"unknown field falls back", "evil; DROP TABLE tasks", "ORDER BY created_at DESC"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, mock := newRepoWithMock(t)
			f := ListFilter{Page: 1, Limit: 10, Sort: tt.sort}

			mock.ExpectQuery(`^SELECT COUNT\(\*\) FROM tasks WHERE deleted_at IS NULL$`).
				WithArgs().
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			mock.ExpectQuery("^SELECT id, title, description, status, assignee, due_date, created_at, updated_at FROM tasks WHERE deleted_at IS NULL "+tt.expected+" LIMIT \\? OFFSET \\?$").
				WithArgs(10, 0).
				WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "status", "assignee", "due_date", "created_at", "updated_at"}))

			if _, _, err := repo.List(context.Background(), f); err != nil {
				t.Fatalf("list: %v", err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet expectations: %v", err)
			}
		})
	}
}

// Task 4: a UNIQUE(title) violation (MySQL 1062) must surface as
// ErrDuplicateTitle, not as a generic error.
func TestCreateMapsDuplicateKeyToDomainError(t *testing.T) {
	repo, mock := newRepoWithMock(t)
	mock.ExpectExec(`INSERT INTO tasks \(title, description, status, assignee, due_date\) VALUES \(\?, \?, \?, \?, \?\)$`).
		WillReturnError(&mysql.MySQLError{Number: 1062, Message: "Duplicate entry 'dup' for key 'tasks.uq_tasks_title'"})

	_, err := repo.Create(context.Background(), &Task{Title: "dup", Status: StatusTodo})
	if !errors.Is(err, ErrDuplicateTitle) {
		t.Fatalf("err = %v, want ErrDuplicateTitle", err)
	}
}

func TestCreateStoresTaskAndSetsID(t *testing.T) {
	repo, mock := newRepoWithMock(t)
	mock.ExpectExec(`INSERT INTO tasks \(title, description, status, assignee, due_date\) VALUES \(\?, \?, \?, \?, \?\)$`).
		WithArgs("first", nil, StatusTodo, nil, nil).
		WillReturnResult(sqlmock.NewResult(7, 1))
	mock.ExpectQuery(`^SELECT id, title, description, status, assignee, due_date, created_at, updated_at FROM tasks WHERE id = \? AND deleted_at IS NULL$`).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "status", "assignee", "due_date", "created_at", "updated_at"}).
			AddRow(7, "first", nil, StatusTodo, nil,
				time.Date(2026, 9, 27, 0, 0, 0, 0, time.UTC),
				time.Date(2026, 9, 27, 10, 0, 0, 0, time.UTC),
				time.Date(2026, 9, 27, 10, 0, 0, 0, time.UTC)))

	created, err := repo.Create(context.Background(), &Task{Title: "first", Status: StatusTodo})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID != 7 || created.UpdatedAt.IsZero() {
		t.Fatalf("created = %+v, want the re-selected row with id 7 and real timestamps", created)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// Task 5: update — updating a missing or already-deleted row is 404 territory.
func TestUpdateReturnsNotFoundWhenNoRowUpdated(t *testing.T) {
	repo, mock := newRepoWithMock(t)
	mock.ExpectExec(`UPDATE tasks SET title = \?, description = \?, status = \?, assignee = \?, due_date = \? WHERE id = \? AND deleted_at IS NULL$`).
		WithArgs("renamed", nil, StatusTodo, nil, nil, 99).
		WillReturnResult(sqlmock.NewResult(0, 0))

	_, err := repo.Update(context.Background(), &Task{ID: 99, Title: "renamed", Status: StatusTodo})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestUpdateReturnsFreshRow(t *testing.T) {
	repo, mock := newRepoWithMock(t)
	mock.ExpectExec(`UPDATE tasks SET title = \?, description = \?, status = \?, assignee = \?, due_date = \? WHERE id = \? AND deleted_at IS NULL$`).
		WithArgs("renamed", nil, StatusTodo, nil, nil, 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(`^SELECT id, title, description, status, assignee, due_date, created_at, updated_at FROM tasks WHERE id = \? AND deleted_at IS NULL$`).
		WithArgs(1).
		WillReturnRows(oneTaskRow())

	updated, err := repo.Update(context.Background(), &Task{ID: 1, Title: "renamed", Status: StatusTodo})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.ID != 1 || updated.Title != "first" {
		t.Fatalf("updated = %+v, want the re-selected row", updated)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSoftDelete(t *testing.T) {
	t.Run("soft deletes the active row", func(t *testing.T) {
		repo, mock := newRepoWithMock(t)
		mock.ExpectExec(`^UPDATE tasks SET deleted_at = NOW\(\) WHERE id = \? AND deleted_at IS NULL$`).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(1, 1))

		if err := repo.SoftDelete(context.Background(), 1); err != nil {
			t.Fatalf("delete: %v", err)
		}
	})
	t.Run("already deleted or missing is not found", func(t *testing.T) {
		repo, mock := newRepoWithMock(t)
		mock.ExpectExec(`^UPDATE tasks SET deleted_at = NOW\(\) WHERE id = \? AND deleted_at IS NULL$`).
			WithArgs(99).
			WillReturnResult(sqlmock.NewResult(0, 0))

		if err := repo.SoftDelete(context.Background(), 99); !errors.Is(err, ErrNotFound) {
			t.Fatalf("err = %v, want ErrNotFound", err)
		}
	})
}
