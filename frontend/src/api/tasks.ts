import {request} from './client';
import type {
  ListResponse,
  ListTasksQuery,
  Task,
  TaskDataResponse,
  UpdateTaskInput,
} from '../types/task';

/** GET /api/tasks — list + meta untuk pagination. Satu-satunya endpoint list FE. */
export function listTasks(
  query: ListTasksQuery = {},
  signal?: AbortSignal,
): Promise<ListResponse> {
  return request<ListResponse>('/api/tasks', {query: {...query}, signal});
}

/**
 * PUT /api/tasks/:id — full update (`title` + `status` wajib).
 * BE membalas envelope `{data: Task}` → di-unwrap di sini supaya pemakai dapat Task.
 */
export function updateTask(id: number, input: UpdateTaskInput): Promise<Task> {
  return request<TaskDataResponse>(`/api/tasks/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  }).then(res => res.data);
}
