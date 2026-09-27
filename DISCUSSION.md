# Diskusi Teknis — Fullstack Engineer Assessment (Task Management)

> Dokumen diskusi, bukan implementasi. Berisi analisa task, struktur folder yang diusulkan,
> keputusan teknis yang perlu disepakati, dan best practice yang akan dipakai.

---

## 1. Analisa Assessment & Prioritas

### Mapping bobot evaluasi → fokus kerja

| Kriteria | Bobot | Artinya secara praktis |
|---|---|---|
| Go | 35% | Arsitektur berlapis yang rapi, idiomatic Go, `context`, error handling, validasi input |
| Frontend | 20% | Componentization, hooks, UX states (loading/error/empty) |
| Redis | 15% | Desain cache key, TTL, invalidation, graceful degradation |
| SQL | 10% | Skema + index + migration + dynamic query yang aman (anti SQL injection) |
| Testing | 10% | Unit test backend (update/search/cache) + minimal 1 component test frontend |
| Code Quality | 5% | Linter, formatter, konvensi penamaan konsisten |
| Documentation | 5% | README lengkap: setup, env, API contract, cara run test |

**Kesimpulan strategi:** backend dan Redis adalah 50% nilai → kerjakan paling awal dan paling rapi.
Frontend solid tapi tidak perlu over-engineer. Testing jangan dikerjakan di akhir — kerjakan
bersamaan per fitur supaya tidak tergesa.

### Catatan kondisi repo saat ini

- Repo praktis **kosong** — hanya `go.mod` + `go.sum` (Gin, `go-sql-driver/mysql`, `go-redis` sudah ada). Jadi ini greenfield; "existing functionality" yang dimaksud soal adalah aplikasi yang kita bangun sendiri di awal (CRUD dasar), lalu fitur assessment ditambahkan di atasnya tanpa merusaknya.
- `.git` saat ini berada **di dalam folder `backend/`**. Deliverable mencakup frontend → sebaiknya dijadikan **monorepo** dengan git root di `task-management-assessment/` (perlu re-init / pindah `.git` satu level ke atas). Perlu konfirmasi.
- Di `go.mod` ada dependensi `// indirect` yang aneh (`mongo-driver`, `quic-go`) — nanti dibersihkan dengan `go mod tidy` saat mulai coding.

---

## 2. Keputusan Teknis yang Perlu Didiskusikan (Open Questions)

Ini pertanyaan yang menurut saya layak dikonfirmasi sebelum mulai coding (atau minimal
kita sepakati asumsinya dan tulis di README):

1. **Semantik `PUT`** — full update (semua field required) atau partial update (semacam PATCH)?
   Usulan: `PUT` = full update dengan validasi field wajib (`title`, `status`), sesuai semantik HTTP;
   field opsional (`description`, `assignee`, `due_date`) boleh di-null-kan. Ini paling mudah
   dipertanggungjawabkan di majelis review.
2. **Unique title vs soft delete** — kalau `title` di-UNIQUE plain, judul yang sudah soft-deleted
   akan "mengunci" judul itu selamanya. Opsi dibahas di §5 (generated column). Rekomendasi saya: aktif
   yang unik, yang terhapus bebas.
3. **Parameter `sort`** — field apa saja yang boleh? Usulan whitelist: `title`, `status`, `assignee`,
   `created_at`, `updated_at` + arah `asc|desc` (param terpisah `order`). ORDER BY tidak bisa
   di-parameterize → wajib whitelist agar bebas SQL injection.
4. **Filter `assignee`** — exact match atau partial (LIKE)? Usulan: exact match (karena di UI nanti
   berupa pilihan), `keyword` yang pakai LIKE untuk `title`.
5. **Strategi invalidasi cache** — hapus semua key dengan prefix `tasks:list:*` via `SCAN`
   (sederhana, selalu benar) vs hapus presisi per kombinasi query (rumit, rawan bocor).
   Usulan: flush per prefix. Cukup untuk skala ini, dan gampang di-test.
6. **State management frontend** — plain hooks (`useState`/`useEffect` + axios) vs React Query?
   Usulan: plain hooks (tanpa dependency tambahan, sesuai durasi 1–2 hari); di README disebutkan
   bahwa di production kita akan pakai TanStack Query untuk caching/retry/refetch otomatis.
7. **Monorepo vs repo terpisah** — usulan monorepo (`backend/` + `frontend/`), perlu pindahkan git root.
8. **Out of scope (ditulis sebagai asumsi di README):** tidak ada auth, tidak ada tabel `users`
   (`assignee` cukup string), tidak ada websockets/realtime.

---

## 3. Struktur Folder yang Diusulkan (Monorepo)

```
task-management-assessment/
├── README.md                        # dokumentasi utama (setup, API, test)
├── docker-compose.yml               # MySQL 8 + Redis 7 (+ optional adminer)
├── DISCUSSION.md                    # dokumen ini (opsional dihapus saat submit)
├── Makefile                         # shortcut: run, test, migrate, lint, docker
│
├── backend/                         # Go (Gin) + MySQL + Redis
│   ├── cmd/
│   │   └── api/
│   │       └── main.go              # entrypoint: load config → koneksi DB/Redis → wiring → graceful shutdown
│   ├── internal/                    # tidak bisa diimport dari luar module (Go convention)
│   │   ├── config/
│   │   │   └── config.go            # baca env (12-factor), default value, validasi config
│   │   ├── database/
│   │   │   ├── mysql.go             # *sql.DB + pool setting (MaxOpenConns, ConnMaxLifetime)
│   │   │   └── redis.go             # redis.UniversalClient init + health check
│   │   ├── middleware/
│   │   │   ├── logger.go            # request logging (method, path, status, latency)
│   │   │   ├── recovery.go          # panic → 500 terformat (jangan pakai bawaan gin yang mentah)
│   │   │   ├── cors.go
│   │   │   └── error.go             # PUSAT pemetaan error domain → HTTP response terformat
│   │   └── task/                    # satu module domain = semua layer untuk "task"
│   │       ├── handler.go           # HTTP layer: bind & validate request, panggil service, tulis response
│   │       ├── handler_test.go
│   │       ├── service.go           # business logic + orkestrasi cache
│   │       ├── service_test.go      # termasuk test cache invalidation (miniredis)
│   │       ├── repository.go        # SQL murni: filtering, pagination, soft delete
│   │       ├── repository_test.go   # search & soft-delete (sqlmock / MySQL docker)
│   │       ├── model.go             # entity Task (plain struct, tag db)
│   │   │   ├── dto.go               # request/response structs (tag binding+json terpisah)
│   │   │   ├── cache.go             # cache key builder, Get/Set/Invalidate (interface → mockable)
│   │   │   └── errors.go            # error domain: ErrNotFound, ErrDuplicateTitle, ErrValidation
│   │   └── apperr/                  # (opsional) error envelope umum lintas domain
│   ├── migrations/
│   │   ├── 000001_create_tasks.up.sql
│   │   ├── 000001_create_tasks.down.sql
│   │   └── 000002_add_indexes.up.sql (jika perlu)
│   ├── .env.example
│   ├── Dockerfile                   # multi-stage build (builder → distroless/alpine)
│   └── .golangci.yml
│
└── frontend/                        # React Native + TypeScript
    ├── src/
    │   ├── api/
    │   │   ├── client.ts            # axios instance: baseURL, timeout, interceptor error
    │   │   └── tasks.ts             # fungsi listTasks/getTask/updateTask/deleteTask (typed)
    │   ├── components/
    │   │   ├── SearchInput.tsx      # input + debounce internal
    │   │   ├── StatusFilter.tsx     # chips/dropdown: all | todo | in_progress | done
    │   │   ├── TaskItem.tsx         # 1 baris task + tombol edit/delete
    │   │   ├── TaskList.tsx         # FlatList + empty state
    │   │   ├── Pagination.tsx       # prev/next + info halaman
    │   │   ├── EditTaskModal.tsx    # form + validasi + loading saat submit
    │   │   └── common/              # Button, LoadingOverlay, ErrorBanner, SkeletonList
    │   ├── screens/
    │   │   └── TaskListScreen.tsx   # komposisi semua komponen di atas
    │   ├── hooks/
    │   │   └── useTasks.ts          # state: list, meta, loading, error; refetch; AbortController
    │   ├── types/
    │   │   └── task.ts              # type Task, TaskListMeta, enum status
    │   ├── theme/                   # warna, spacing, typography (konsistensi UI)
    │   └── __tests__/
    │       └── SearchInput.test.tsx # minimal 1: search / task list (Jest + RNTL)
    ├── App.tsx
    └── jest.config.js
```

**Alasan struktur ini:**
- `internal/` → idiom Go untuk kode private module.
- Layer per-domain (`task/handler|service|repository`) bukan per-layer global — dependency mengalir satu arah: `handler → service → repository`, tidak pernah kebalikan. `service` menerima interface `TaskRepository` dan `TaskCache` → mudah di-mock saat test (kunci untuk nilai Testing 10%).
- `dto.go` dipisah dari `model.go` → kontrak API tidak dempet dengan skema DB (tag `binding` + `json` berbeda kebutuhan).
- Migrasi sebagai file SQL versioned (golang-migrate) → deliverable "DB migration" terpenuhi dan bisa di-review.
- Frontend: logika data terpusat di `useTasks`, komponen presentational → gampang dites.

---

## 4. Kontrak API & Error Response

### Endpoint

| Method | Path | Keterangan |
|---|---|---|
| GET | `/api/tasks` | list + filter `status`, `keyword`, `assignee`, `page`, `limit`, `sort`, `order` |
| GET | `/api/tasks/:id` | detail (404 jika tidak ada / sudah soft-deleted) |
| POST | `/api/tasks` | create (409 jika title duplikat) |
| PUT | `/api/tasks/:id` | update full (404 / 400 / 409) |
| DELETE | `/api/tasks/:id` | soft delete → 204 (404 jika tidak ada) |

### Envelope sukses (konsisten)

```json
// GET /api/tasks?status=todo&page=1&limit=10
{
  "data": [ { "id": 1, "title": "...", "status": "todo", "assignee": "...", "...": "..." } ],
  "meta": { "page": 1, "limit": 10, "total_items": 42, "total_pages": 5 }
}
```

### Envelope error (konsisten, dipusatkan di middleware `error.go`)

```json
{
  "error": {
    "code": "DUPLICATE_TITLE",            // kode stabil, mesin-membaca
    "message": "task title already exists" // pesan manusia
  }
}
```

Pemetaan error domain → HTTP (tipe error khusus, dicek dengan `errors.As`):

| Error domain | HTTP | Kode |
|---|---|---|
| `ErrValidation` | 400 | `VALIDATION_ERROR` (+ detail per field) |
| `ErrNotFound` | 404 | `NOT_FOUND` |
| `ErrDuplicateTitle` | 409 | `DUPLICATE_TITLE` |
| error tak terduga | 500 | `INTERNAL_ERROR` (stack hanya ke log, tidak bocor ke client) |

Handler **tidak pernah** memanggil `c.JSON(...)` untuk error secara manual — cukup `return err`,
middleware yang memformat. Ini yang membuat "consistent error responses" terjamin dan mudah dites.

---

## 5. Skema DB & Migration (usulan)

```sql
-- 000001_create_tasks.up.sql
CREATE TABLE tasks (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    title       VARCHAR(255) NOT NULL,
    description TEXT NULL,
    status      ENUM('todo','in_progress','done') NOT NULL DEFAULT 'todo',
    assignee    VARCHAR(255) NULL,
    due_date    DATE NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at  TIMESTAMP NULL DEFAULT NULL,

    -- unik hanya untuk task AKTIF: baris soft-deleted menjadi NULL → dikecualikan
    -- dari constraint (MySQL mengizinkan banyak NULL di unique index)
    active_title VARCHAR(255) GENERATED ALWAYS AS (
        CASE WHEN deleted_at IS NULL THEN title END
    ) STORED,
    UNIQUE KEY uq_tasks_active_title (active_title),

    KEY idx_tasks_status_deleted (status, deleted_at),
    KEY idx_tasks_assignee_deleted (assignee, deleted_at),
    KEY idx_tasks_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

Poin SQL yang dinilai (10%):
- **Duplicate title → 409, bukan 500**: sumber kebenaran = constraint DB. Repository menangkap
  `*mysql.MySQLError` dengan `Number == 1062` → translate ke `ErrDuplicateTitle` → 409.
  (Cek duplikat manual dengan `SELECT` dulu itu race-prone; constraint + catch 1062 itu benar.)
- **Index composite** mengikuti pola query filter + soft delete.
- **ORDER BY di-whitelist** di service layer (map `sort` param → nama kolom aman) — ORDER BY tidak
  bisa di-placeholder, jadi whitelist adalah satu-satunya cara aman.
- Semua query (list, get, update, delete) **wajib** menambahkan `deleted_at IS NULL` —
  ini bug fix "hide soft-deleted tasks" sekaligus jaga-jaga agar update/delete terhadap task
  yang sudah terhapus mengembalikan 404, bukan meresurrect data.
- Keyword search: `WHERE title LIKE CONCAT('%', ?, '%')` — parameterized, bukan string concat.

---

## 6. Best Practice per Task

### Task 1 — Backend (40%)
- Layered architecture: `handler` (HTTP) → `service` (business logic) → `repository` (SQL). Satu arah, via interface.
- Validasi di boundary: `binding` tags Gin untuk bentuk request + validasi semantik (enum status, range page/limit: `page ≥ 1`, `1 ≤ limit ≤ 100`, default `page=1`, `limit=10`).
- Dynamic WHERE dibangun dengan slice `[]any` + placeholder `?` — tidak pernah string concat nilai user.
- `context.Context` mengalir sampai ke query (`QueryContext`) + `context.WithTimeout` per request.
- Kembalikan `meta` pagination dari `COUNT(*)` terpisah — total tidak boleh mengikuti LIMIT.
- Graceful shutdown di `main.go` (signal → `server.Shutdown` + tutup DB/Redis).

### Task 2 — Redis (15%)
- Cache di service layer (bukan middleware) supaya testable dan invalidation eksplisit.
- **Cache key memuat seluruh query param secara kanonik** (diurutkan, dinormalisasi):
  ```
  tasks:list:assignee=budi&keyword=fix&limit=10&page=1&order=desc&sort=created_at&status=todo
  ```
  Builder key: parse query → struct → serialize field-by-field terurut → duplikat key yang sama
  dari query yang hanya beda urutan param tetap HIT.
- TTL 60 detik (`SET key json EX 60`).
- Value = envelope lengkap (`data` + `meta`) → hit juga menghemat query COUNT.
- Invalidasi: setelah POST/PUT/DELETE sukses → `SCAN` prefix `tasks:list:*` + `DEL`
  (`SCAN`, **bukan** `KEYS` — KEYS memblokir Redis di production).
- **Graceful degradation**: kalau Redis down, jangan 500 — log warning dan lanjut ke DB.
  Cache adalah optimasi, bukan dependency fungsional.
- Catatan di README: untuk beban tinggi, stampede bisa dicegah dengan `singleflight` (disebut, tidak wajib diimplementasi).

### Task 3 — Frontend (25%)
- Semua akses API lewat `src/api/` (typed, satu axios instance) — komponen tidak pernah fetch mentah.
- `useTasks` hook tunggal: state `data | loading | error`, `refetch()`, `AbortController` untuk
  membatalkan request basi saat user mengetik cepat (cegah race condition hasil lama menimpa baru).
- Search di-debounce ±300–500ms; filter status & pagination memicu fetch langsung; ganti filter → reset `page` ke 1.
- Loading: skeleton list saat fetch awal/ganti halaman; spinner + disable tombol saat submit edit;
  error banner dengan tombol retry; empty state.
- `EditTaskModal`: form terkontrol, validasi sisi client (title wajib, status enum), tampilkan pesan
  409 dari server, tutup modal → `refetch()` (sekaligus bug fix "refresh list after update").
- `FlatList` (bukan `.map` di View) untuk performa render.

### Task 4 — Bug Fixes (10%)
- 409 duplicate: catch `1062` dari constraint `uq_tasks_active_title` (lihat §5).
- Refresh list after update: `refetch()` setelah PUT sukses di frontend + invalidasi cache di backend (keduanya — hanya salah satu masih bisa menampilkan data basi dari cache).
- Hide soft-deleted: `deleted_at IS NULL` di SEMUA query + test khusus yang membuktikannya.

### Task 5 — Testing (10%)
Backend (Go):
- `handler_test.go` — `httptest` + gin test mode, mock service → test PUT sukses, 404, 409, 400.
- `service_test.go` — mock repository + **miniredis** (`alicebob/miniredis`) → test cache invalidation:
  GET → hit DB; SET key; PUT → key hilang dari miniredis; GET berikutnya → hit DB lagi.
- `repository_test.go` — sqlmock (atau MySQL via docker-compose) → test search (WHERE dinamis benar),
  soft delete tersembunyi, pagination meta.

Frontend:
- Jest + React Native Testing Library. Minimal satu: `SearchInput` (mengetik → debounce → onChange
  terpanggil sekali, fake timers) atau `TaskList` (render item, empty state, loading).

Prinsip: test menyusuri **interface**, bukan implementasi → refactor aman.

---

## 7. Tooling & Code Quality (5%)

- Backend: `gofmt`/`goimports`, `golangci-lint` (errcheck, govet, staticcheck), konvensi nama Go standar.
- Frontend: TypeScript strict, ESLint + Prettier.
- Konfigurasi via env (`.env.example` dicommit, `.env` di-gitignore) — 12-factor.
- `docker-compose.yml` untuk MySQL + Redis supaya reviewer jalan dengan satu perintah.
- `Makefile`: `make run`, `make test`, `make migrate-up`, `make lint`.

## 8. Strategi Git & Commit (deliverable "git repository" ikut dinilai)

Conventional commits, commit kecil per fitur — riwayat yang bisa dibaca reviewer:

```
chore(backend): scaffold project, config, docker-compose
feat(db): add tasks migration with soft delete + active title unique constraint
feat(api): task CRUD endpoints with consistent error envelope
feat(api): list filtering, pagination, and sorting
fix(api): map duplicate title to 409 instead of 500
feat(cache): cache GET /api/tasks for 60s with query-param keys
feat(cache): invalidate list cache on create/update/delete
feat(frontend): task list screen with search, status filter, pagination
feat(frontend): edit modal with loading state and list refresh
test(backend): service, handler, repository, and cache invalidation tests
test(frontend): SearchInput component test
docs: README with setup, API contract, and test instructions
```

## 9. Rencana Eksekusi (1–2 hari)

| Tahap | Isi |
|---|---|
| 0. Setup (±1 jam) | Monorepo, docker-compose, config, koneksi DB/Redis, migration, CRUD dasar jalan |
| 1. Backend inti | Filtering/pagination/sort + PUT + soft delete + error envelope (Task 1 & 4) |
| 2. Redis | Cache + invalidation (Task 2) |
| 3. Test backend | Ditulis per fitur di tahap 1–2, bukan di akhir (Task 5) |
| 4. Frontend | List + filter + pagination + modal + loading (Task 3) + 1 component test |
| 5. Polish | README, .env.example, lint bersih, review ulang requirement checklist per item |

**Checklist akhir**: cocokkan satu-per-satu setiap bullet task 1–5 dengan implementasi —
assessment dinilai per bullet, jadi pastikan tidak ada yang terlewat.
