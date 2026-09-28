package task

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// fakeRepo is a hand-written Repository double that counts calls, so tests
// can assert whether a read actually reached the database or the cache.
type fakeRepo struct {
	listFn   func(ctx context.Context, f ListFilter) ([]Task, int, error)
	createFn func(ctx context.Context, t *Task) (*Task, error)
	updateFn func(ctx context.Context, t *Task) (*Task, error)
	deleteFn func(ctx context.Context, id int64) error

	listCalls, createCalls, updateCalls, deleteCalls int
}

func (r *fakeRepo) List(ctx context.Context, f ListFilter) ([]Task, int, error) {
	r.listCalls++
	return r.listFn(ctx, f)
}

func (r *fakeRepo) Create(ctx context.Context, t *Task) (*Task, error) {
	r.createCalls++
	return r.createFn(ctx, t)
}

func (r *fakeRepo) Update(ctx context.Context, t *Task) (*Task, error) {
	r.updateCalls++
	return r.updateFn(ctx, t)
}

func (r *fakeRepo) SoftDelete(ctx context.Context, id int64) error {
	r.deleteCalls++
	return r.deleteFn(ctx, id)
}

func newTestCache(t *testing.T) (TaskCache, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })
	return NewRedisCache(rdb, time.Minute), mr
}

func sampleTask(id int64) Task {
	return Task{ID: id, Title: "first", Status: StatusTodo}
}

// TestListCachesSecondRead (Task 5: cache): the first list read fills the
// cache, the identical second read must not reach the repository.
func TestListCachesSecondRead(t *testing.T) {
	cache, mr := newTestCache(t)
	repo := &fakeRepo{listFn: func(ctx context.Context, f ListFilter) ([]Task, int, error) {
		return []Task{sampleTask(1)}, 1, nil
	}}
	svc := NewService(repo, cache)
	ctx := context.Background()

	first, err := svc.List(ctx, ListTasksQuery{})
	if err != nil {
		t.Fatalf("first list: %v", err)
	}
	if repo.listCalls != 1 {
		t.Fatalf("after first read: list calls = %d, want 1", repo.listCalls)
	}

	second, err := svc.List(ctx, ListTasksQuery{})
	if err != nil {
		t.Fatalf("second list: %v", err)
	}
	if repo.listCalls != 1 {
		t.Fatalf("second read hit the repository (list calls = %d); cache miss?", repo.listCalls)
	}
	if len(second.Data) != 1 || second.Data[0].ID != 1 {
		t.Fatalf("cached response = %+v, want the cached task", second)
	}
	if first.Meta.TotalItems != 1 {
		t.Fatalf("meta total items = %d, want 1", first.Meta.TotalItems)
	}

	q := ListTasksQuery{}
	normalizeListQuery(&q)
	key := ListCacheKey(q)
	if !mr.Exists(key) {
		t.Fatalf("cache key %q not stored", key)
	}
	if ttl := mr.TTL(key); ttl <= 0 || ttl > time.Minute {
		t.Fatalf("ttl for %q = %v; want within the 60s TTL", key, ttl)
	}
}

// TestUpdateInvalidatesListCache (Task 5: cache invalidation): after a
// successful update, the next identical list read must reach the repository.
func TestUpdateInvalidatesListCache(t *testing.T) {
	cache, _ := newTestCache(t)
	repo := &fakeRepo{
		listFn: func(ctx context.Context, f ListFilter) ([]Task, int, error) {
			return []Task{sampleTask(1)}, 1, nil
		},
		updateFn: func(ctx context.Context, tsk *Task) (*Task, error) {
			updated := *tsk
			updated.Title = "renamed"
			return &updated, nil
		},
	}
	svc := NewService(repo, cache)
	ctx := context.Background()

	if _, err := svc.List(ctx, ListTasksQuery{}); err != nil {
		t.Fatalf("list: %v", err)
	}
	if _, err := svc.Update(ctx, 1, UpdateTaskRequest{Title: "renamed", Status: StatusTodo}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if _, err := svc.List(ctx, ListTasksQuery{}); err != nil {
		t.Fatalf("list after update: %v", err)
	}
	if repo.listCalls != 2 {
		t.Fatalf("list calls = %d, want 2 (cache must be invalidated by update)", repo.listCalls)
	}
}

// TestDeleteInvalidatesListCache mirrors the update case for soft delete.
func TestDeleteInvalidatesListCache(t *testing.T) {
	cache, _ := newTestCache(t)
	repo := &fakeRepo{
		listFn: func(ctx context.Context, f ListFilter) ([]Task, int, error) {
			return []Task{sampleTask(1)}, 1, nil
		},
		deleteFn: func(ctx context.Context, id int64) error { return nil },
	}
	svc := NewService(repo, cache)
	ctx := context.Background()

	if _, err := svc.List(ctx, ListTasksQuery{}); err != nil {
		t.Fatalf("list: %v", err)
	}
	if err := svc.Delete(ctx, 1); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := svc.List(ctx, ListTasksQuery{}); err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if repo.listCalls != 2 {
		t.Fatalf("list calls = %d, want 2 (cache must be invalidated by delete)", repo.listCalls)
	}
}

// Task 5: update behaviour — success path is covered above; here the error
// paths (not found, duplicate title) must pass through unchanged.
func TestUpdateErrorPaths(t *testing.T) {
	tests := []struct {
		name    string
		repoErr error
		want    error
	}{
		{"not found", ErrNotFound, ErrNotFound},
		{"duplicate title", ErrDuplicateTitle, ErrDuplicateTitle},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, _ := newTestCache(t)
			repo := &fakeRepo{updateFn: func(ctx context.Context, tsk *Task) (*Task, error) {
				return nil, tt.repoErr
			}}
			svc := NewService(repo, cache)

			_, err := svc.Update(context.Background(), 1, UpdateTaskRequest{Title: "x", Status: StatusTodo})
			if !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestCreateDuplicateTitle(t *testing.T) {
	cache, _ := newTestCache(t)
	repo := &fakeRepo{createFn: func(ctx context.Context, tsk *Task) (*Task, error) {
		return nil, ErrDuplicateTitle
	}}
	svc := NewService(repo, cache)

	_, err := svc.Create(context.Background(), CreateTaskRequest{Title: "dup"})
	if !errors.Is(err, ErrDuplicateTitle) {
		t.Fatalf("err = %v, want ErrDuplicateTitle", err)
	}
}

func TestNormalizeListQuery(t *testing.T) {
	tests := []struct {
		name string
		in   ListTasksQuery
		want ListTasksQuery
	}{
		{"defaults", ListTasksQuery{}, ListTasksQuery{Page: 1, Limit: 10, Sort: defaultSort}},
		{"clamps limit", ListTasksQuery{Limit: 500}, ListTasksQuery{Page: 1, Limit: maxLimit, Sort: defaultSort}},
		{"negative page", ListTasksQuery{Page: -2, Limit: -5}, ListTasksQuery{Page: 1, Limit: 10, Sort: defaultSort}},
		{"keeps valid values", ListTasksQuery{Page: 3, Limit: 5, Sort: "title:asc", Status: StatusDone, Keyword: "fix", Assignee: "budi"},
			ListTasksQuery{Page: 3, Limit: 5, Sort: "title:asc", Status: StatusDone, Keyword: "fix", Assignee: "budi"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.in
			normalizeListQuery(&got)
			if got != tt.want {
				t.Fatalf("normalize = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// Task 2 robustness: a Redis failure must degrade to a plain DB read,
// never fail the request.
func TestListDegradesWhenRedisIsDown(t *testing.T) {
	cache, mr := newTestCache(t)
	repo := &fakeRepo{listFn: func(ctx context.Context, f ListFilter) ([]Task, int, error) {
		return []Task{sampleTask(1)}, 1, nil
	}}
	svc := NewService(repo, cache)

	mr.Close() // simulate Redis going down

	resp, err := svc.List(context.Background(), ListTasksQuery{})
	if err != nil {
		t.Fatalf("list with redis down: %v", err)
	}
	if repo.listCalls != 1 || len(resp.Data) != 1 {
		t.Fatalf("expected a DB read with redis down: calls=%d data=%+v", repo.listCalls, resp.Data)
	}
}

// Task 2: the cache key must include every query parameter and be stable.
func TestListCacheKeyIncludesAllParams(t *testing.T) {
	base := ListTasksQuery{Status: StatusTodo, Keyword: "fix", Assignee: "budi", Page: 2, Limit: 5, Sort: "title:asc"}

	a := ListCacheKey(base)
	b := ListCacheKey(base)
	if a != b {
		t.Fatalf("same query produced different keys:\n%s\n%s", a, b)
	}
	for _, part := range []string{"status=todo", "keyword=fix", "assignee=budi", "page=2", "limit=5", "sort=title%3Aasc"} {
		if !strings.Contains(a, part) {
			t.Fatalf("key %q missing %q", a, part)
		}
	}

	different := base
	different.Page = 3
	if ListCacheKey(different) == a {
		t.Fatal("changing a query parameter must change the cache key")
	}
	if !strings.HasPrefix(a, listKeyPrefix) {
		t.Fatalf("key %q must live under prefix %q", a, listKeyPrefix)
	}
}
