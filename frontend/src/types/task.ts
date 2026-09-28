/**
 * Tipe yang mencerminkan kontrak BE (`backend/internal/task/dto.go`).
 * Field opsional BE (`*string`) menjadi `| null` di sini, karena JSON-nya bisa `null`.
 */

/** Kolom `status` BE: binding `oneof=todo in_progress done`. */
export type Status = 'todo' | 'in_progress' | 'done';

/** `TaskResponse` BE — dipakai untuk list sekaligus pre-fill edit modal. */
export interface Task {
  id: number;
  title: string;
  description: string | null;
  status: Status;
  assignee: string | null;
  /** Format `YYYY-MM-DD` (BE memformat sebelum kirim), null kalau kosong. */
  due_date: string | null;
  /** RFC3339 dari `time.Time` BE. */
  created_at: string;
  updated_at: string;
}

/** `ListMeta` BE — sumber pagination FE (Fase FE-3). */
export interface ListMeta {
  page: number;
  limit: number;
  total_items: number;
  total_pages: number;
}

/** Envelope `GET /api/tasks`: `{data, meta}`. */
export interface ListResponse {
  data: Task[];
  meta: ListMeta;
}

/** Query string `GET /api/tasks`. Semua opsional; BE yang menormalkan default. */
export interface ListTasksQuery {
  status?: Status;
  keyword?: string;
  assignee?: string;
  page?: number;
  limit?: number;
  /** Format BE: `"field:asc"` | `"field:desc"`. */
  sort?: string;
}

/** Body `PUT /api/tasks/:id` — full update: `title` + `status` wajib (`UpdateTaskRequest` BE). */
export interface UpdateTaskInput {
  title: string;
  status: Status;
  description?: string | null;
  assignee?: string | null;
  due_date?: string | null;
}

/** Envelope `PUT`: `{data: Task}` (`TaskData` BE). */
export interface TaskDataResponse {
  data: Task;
}

/** Isi `error` dari envelope error BE: `{error: {code, message, fields?}}` (middleware/error.go). */
export interface ApiErrorBody {
  code: string;
  message: string;
  /** Hanya ada pada 400 `VALIDATION_ERROR`. */
  fields?: Record<string, string>;
}
