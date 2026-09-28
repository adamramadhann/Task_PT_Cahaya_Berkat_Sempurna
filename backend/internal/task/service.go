package task

import (
	"context"
	"time"
)

const (
	defaultPage  = 1
	defaultLimit = 10
	maxLimit     = 100
	defaultSort  = "created_at:desc"
)

// Service is the business layer used by the HTTP handler.
type Service interface {
	List(ctx context.Context, q ListTasksQuery) (*ListResponse, error)
	Create(ctx context.Context, req CreateTaskRequest) (TaskResponse, error)
	Update(ctx context.Context, id int64, req UpdateTaskRequest) (TaskResponse, error)
	Delete(ctx context.Context, id int64) error
}

type service struct {
	repo  Repository
	cache TaskCache // nil-safe: only a cache, never a dependency
}

// NewService wires the task service. List reads go through the cache;
// every successful mutation invalidates the list cache.
func NewService(repo Repository, cache TaskCache) Service {
	return &service{repo: repo, cache: cache}
}

func (s *service) List(ctx context.Context, q ListTasksQuery) (*ListResponse, error) {
	normalizeListQuery(&q)
	key := ListCacheKey(q)
	if s.cache != nil {
		if resp, ok := s.cache.GetList(ctx, key); ok {
			return resp, nil
		}
	}

	items, total, err := s.repo.List(ctx, ListFilter{
		Status:   q.Status,
		Keyword:  q.Keyword,
		Assignee: q.Assignee,
		Page:     q.Page,
		Limit:    q.Limit,
		Sort:     q.Sort,
	})
	if err != nil {
		return nil, err
	}

	resp := buildListResponse(items, total, q)
	if s.cache != nil {
		s.cache.SetList(ctx, key, resp)
	}
	return resp, nil
}

func (s *service) Create(ctx context.Context, req CreateTaskRequest) (TaskResponse, error) {
	t, err := newTaskFromCreate(req)
	if err != nil {
		return TaskResponse{}, err
	}
	created, err := s.repo.Create(ctx, t)
	if err != nil {
		return TaskResponse{}, err
	}
	if s.cache != nil {
		s.cache.InvalidateList(ctx)
	}
	return newTaskResponse(*created), nil
}

func (s *service) Update(ctx context.Context, id int64, req UpdateTaskRequest) (TaskResponse, error) {
	t, err := newTaskFromUpdate(req)
	if err != nil {
		return TaskResponse{}, err
	}
	t.ID = id
	updated, err := s.repo.Update(ctx, t)
	if err != nil {
		return TaskResponse{}, err
	}
	if s.cache != nil {
		s.cache.InvalidateList(ctx)
	}
	return newTaskResponse(*updated), nil
}

func (s *service) Delete(ctx context.Context, id int64) error {
	if err := s.repo.SoftDelete(ctx, id); err != nil {
		return err
	}
	if s.cache != nil {
		s.cache.InvalidateList(ctx)
	}
	return nil
}

func normalizeListQuery(q *ListTasksQuery) {
	if q.Page < 1 {
		q.Page = defaultPage
	}
	if q.Limit < 1 {
		q.Limit = defaultLimit
	}
	if q.Limit > maxLimit {
		q.Limit = maxLimit
	}
	if q.Sort == "" {
		q.Sort = defaultSort
	}
}

func newTaskFromCreate(req CreateTaskRequest) (*Task, error) {
	status := req.Status
	if status == "" {
		status = StatusTodo
	}
	due, err := parseDueDate(req.DueDate)
	if err != nil {
		return nil, err
	}
	return &Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      status,
		Assignee:    req.Assignee,
		DueDate:     due,
	}, nil
}

func newTaskFromUpdate(req UpdateTaskRequest) (*Task, error) {
	due, err := parseDueDate(req.DueDate)
	if err != nil {
		return nil, err
	}
	return &Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Assignee:    req.Assignee,
		DueDate:     due,
	}, nil
}

func parseDueDate(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil, &ValidationError{Fields: map[string]string{"due_date": "must be YYYY-MM-DD"}}
	}
	return &t, nil
}

func buildListResponse(items []Task, total int, q ListTasksQuery) *ListResponse {
	data := make([]TaskResponse, 0, len(items))
	for _, t := range items {
		data = append(data, newTaskResponse(t))
	}
	return &ListResponse{
		Data: data,
		Meta: ListMeta{
			Page:       q.Page,
			Limit:      q.Limit,
			TotalItems: total,
			TotalPages: (total + q.Limit - 1) / q.Limit,
		},
	}
}
