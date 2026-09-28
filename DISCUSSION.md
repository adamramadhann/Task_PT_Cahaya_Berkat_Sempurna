# Diskusi Teknis — Backend (Fullstack Engineer Assessment: Task Management)

> Fase **Backend**. Frontend ditunda ke appendix. Prinsip rencana ini: **sesuai task, tidak lebih** —
> setiap fase dipetakan ke bullet task, dan ada daftar eksplisit hal yang sengaja TIDAK dikerjakan.

---

## 1. Scoping: In vs Out (anti over-engineering)

### IN — yang dikerjakan (semuanya langsung dari teks soal)

| Item | Sumber di soal |
|---|---|
| `GET /api/tasks` + filter `status, keyword, assignee, page, limit, sort` | Task 1 |
| `PUT /api/tasks/{id}` | Task 1 |
| Soft `DELETE /api/tasks/{id}` | Task 1 |
| Error response konsisten (envelope terpusat) | Task 1 |
| `POST /api/tasks` (create) — **bukan task, tapi prasyarat** | Task 2 menyebut *"invalidate after **create**"*; Task 4 *"duplicate title → 409"* — keduanya hanya ada kalau create ada. Dianggap bagian "existing app" (soal: *joining an existing team*), tapi repo kosong → kita bootstrap sendiri, seminimal mungkin: create + list polos |
| Redis: cache GET list 60s, key memuat query param, invalidate setelah create/update/delete | Task 2 |
| Duplicate title → 409 (bukan 500) | Task 4 |
| Hide soft-deleted di semua query | Task 4 |
| Refresh list setelah update (di BE = invalidasi cache; refetch di FE saat fase frontend) | Task 4 |
| Test backend: update, search, cache invalidation | Task 5 |
| README, migration (up+down), git repo rapi | Deliverables |

### OUT — yang sengaja TIDAK dikerjakan (dan alasannya)

| Tidak dikerjakan | Alasan |
|---|---|
| `GET /api/tasks/:id` | Tidak diminta bullet task manapun; modal edit frontend bisa mengisi form dari data list item. Kalau nanti FE ternyata perlu, tambahannya ~10 baris |
| Full CRUD lengkap / tabel users / auth | Tidak diminta; `assignee` cukup string |
| Dockerfile untuk app Go | Deliverable tidak minta; `docker-compose` cukup untuk MySQL + Redis saja, app jalan via `go run` |
| Custom middleware logger/recovery/CORS | Pakai `gin.Logger()` + `gin.Recovery()` bawaan. CORS pun tidak perlu — React Native tidak punya browser CORS |
| ORM (GORM/sqlx) | `database/sql` + driver (sudah ada di go.mod) cukup dan menunjukkan skill SQL lebih |
| Anti-stampede (singleflight), pub/sub invalidation | Over-engineer untuk soal ini; cukup 1 kalimat di README "future improvements" |
| golangci-lint setup penuh | Cukup `gofmt` + `go vet` (Code Quality hanya 5%). golangci opsional kalau ada waktu |
| Generated column untuk unique title | Default: `UNIQUE(title)` plain + catch `1062` → 409 (persis yang diminta). Varian "unik hanya untuk row aktif" dicatat 1 baris di README sebagai catatan desain, bukan implementasi |

---

## 2. Keputusan Final (asumsi kerja — bisa diveto, tapi tidak menghalangi mulai)

1. **`PUT` = full update**: wajib `title` + `status`; opsional `description`, `assignee`, `due_date` (boleh null).
2. **`sort` satu param**: `sort=created_at:desc` (arah optional, default `created_at:desc`).
   Field di-whitelist: `title | status | assignee | created_at | updated_at` — `ORDER BY` tidak bisa
   di-parameterize, whitelist = satu-satunya cara aman.
3. **Duplicate title** = constraint `UNIQUE(title)` di DB + repository menangkap `1062` → `ErrDuplicateTitle`
   → middleware → 409. (Cek `SELECT` dulu itu race-prone.)
4. **`assignee` filter = exact match**; `keyword` = `LIKE` pada `title`.
5. **Invalidasi cache = hapus semua key prefix `tasks:list:*`** via `SCAN` (bukan `KEYS`). Sederhana, selalu benar.
6. **Redis down = degrade, bukan 500**: error cache di-log warning, request lanjut ke DB.
7. **Validasi list**: `page ≥ 1` (default 1), `1 ≤ limit ≤ 100` (default 10), `status ∈ {todo, in_progress, done}`.

---

## 3. Struktur Folder BE (dipangkas seminimal mungkin)

```
backend/                        # git root saat ini
├── cmd/
│   └── api/
│       └── main.go            # config → koneksi (ping) → wiring router → run + graceful shutdown
├── internal/
│   ├── config/
│   │   └── config.go          # os.Getenv + default (stdlib, tanpa library config)
│   ├── database/
│   │   ├── mysql.go           # *sql.DB + pool + ping saat startup
│   │   └── redis.go           # redis client + ping
│   ├── middleware/
│   │   └── error.go           # SATU-SATUNYA middleware custom: error domain → envelope HTTP
│   └── task/
│       ├── model.go           # entity Task
│       ├── dto.go             # request/response (tag binding + json)
│       ├── errors.go          # ErrNotFound, ErrDuplicateTitle, ErrValidation
│       ├── repository.go      # SQL: filter dinamis, pagination, soft delete, catch 1062
│       ├── service.go         # business logic + orkestrasi cache
│       ├── cache.go           # interface TaskCache + impl Redis (key builder, Get/Set/Invalidate)
│       ├── handler.go         # bind/validate → service → envelope
│       └── *_test.go          # service_test, repository_test (fase 5)
├── migrations/
│   ├── 000001_create_tasks.up.sql
│   └── 000001_create_tasks.down.sql
├── docker-compose.yml         # MySQL 8 + Redis 7 (dengan healthcheck)
├── Makefile                   # run / migrate-up / migrate-down / test
├── .env.example
└── .gitignore                 # .env tidak masuk git
```

Dependency arah satu: `handler → service → repository`. `service` menerima interface
`TaskRepository` + `TaskCache` → mockable → inilah yang membuat test Task 5 mudah.

Alur: `client → gin → handler(bind/validate) → service(logika + cache) → repository(SQL) → MySQL / Redis`
Error dikembalikan sebagai `error` biasa ke atas; `middleware/error.go` yang memformat — handler tidak
pernah menulis envelope error manual.

---

## 4. Kontrak API (4 endpoint — itu saja)

| Method | Path | Keterangan |
|---|---|---|
| GET | `/api/tasks` | list + filter `status, keyword, assignee, page, limit, sort` |
| POST | `/api/tasks` | create (base "existing") — 409 jika title duplikat |
| PUT | `/api/tasks/:id` | full update — 400/404/409 |
| DELETE | `/api/tasks/:id` | soft delete → 204; 404 jika tidak ada/sudah terhapus |

Envelope sukses:
```json
{ "data": [ ... ], "meta": { "page": 1, "limit": 10, "total_items": 42, "total_pages": 5 } }
```
Envelope error (semua endpoint, dipusatkan di `middleware/error.go`):
```json
{ "error": { "code": "DUPLICATE_TITLE", "message": "task title already exists" } }
```

Pemetaan error: `ErrValidation`→400 `VALIDATION_ERROR` · `ErrNotFound`→404 `NOT_FOUND` ·
`ErrDuplicateTitle`→409 `DUPLICATE_TITLE` · lainnya→500 `INTERNAL_ERROR` (detail hanya ke log).

---

## 5. Migration (satu file up + satu down)

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
    UNIQUE KEY uq_tasks_title (title),
    KEY idx_tasks_status_deleted   (status, deleted_at),
    KEY idx_tasks_assignee_deleted (assignee, deleted_at),
    KEY idx_tasks_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
-- down: DROP TABLE tasks;
```

Catatan desain (ditulis di README, bukan diimplementasi): dengan soft delete, `UNIQUE(title)` berarti
judul yang sudah dihapus tetap "terpakai" — kalau mau judul bebas setelah delete, opsinya generated
column `active_title`. Untuk soal ini plain UNIQUE sudah memenuhi ("duplicate → 409").

Poin SQL yang dinilai: index composite mengikuti pola query (filter + `deleted_at IS NULL` + sort),
`LIKE CONCAT('%', ?, '%')` parameterized, whitelist sort, `COUNT(*)` terpisah untuk meta,
`deleted_at IS NULL` di SEMUA query.

---

## 6. RENCANA EKSEKUSI — 7 fase, tiap fase dipetakan ke bullet task

| Fase | Isi | Mengerjakan bagian | Estimasi |
|---|---|---|---|
| 0 | Infra & skeleton | prasyarat | 1 j |
| 1 | Migration + base "existing" (POST + GET polos) + error envelope + fix 409 | Task 4 (409), prasyarat Task 2 | 1,5 j |
| 2 | Filtering + pagination + sort | Task 1 bullet 1 | 1,5 j |
| 3 | PUT + soft DELETE + hide soft-deleted | Task 1 bullet 2–3, Task 4 (hide) | 1,5 j |
| 4 | Redis cache + invalidation | Task 2, Task 4 (refresh via invalidasi) | 1,5 j |
| 5 | Test: update, search, cache invalidation | Task 5 | 1,5 j |
| 6 | README + lint + commit | Deliverables | 1 j |

Total ±9,5 jam → pas untuk durasi 1–2 hari dengan buffer.

### Fase 0 — Infra & skeleton
1. `go mod tidy` (buang dependensi liar: mongo-driver, quic-go).
2. `docker-compose.yml`: MySQL 8 + Redis 7 + healthcheck.
3. `config.go`: env → struct + default (stdlib saja).
4. `mysql.go` / `redis.go`: koneksi + `Ping()` saat startup — gagal koneksi = fail fast.
5. `main.go`: gin default middleware (`Logger`, `Recovery`), router kosong, graceful shutdown.

**DoD:** `docker compose up -d` → `make run` → server jalan, log ping DB & Redis sukses.

### Fase 1 — Migration + base "existing" + fix 409
1. Tulis migration up/down; runner via golang-migrate **CLI** di Makefile (tanpa code tambahan).
2. `model.go`, `dto.go`, `errors.go`.
3. `middleware/error.go` — envelope terpusat sejak awal (dipakai semua endpoint setelahnya).
4. `repository.go`: `Create` (catch `1062` → `ErrDuplicateTitle`) + `List` versi polos (tanpa filter).
5. `handler.go`: `POST /api/tasks`, `GET /api/tasks`; wiring di main.

**DoD:** curl `POST` sukses → 201; `POST` title sama → **409** (bukan 500); `GET` → 200 berisi data.

### Fase 2 — Filtering + pagination + sort (Task 1)
1. `ListTasksQuery` dto: parse + validasi (default page/limit, enum status, whitelist sort).
2. Repository `List`: WHERE dinamis dari slice kondisi + `[]any` (placeholder `?`, tidak pernah concat
   string user), `deleted_at IS NULL` selalu ada, `ORDER BY` dari whitelist, `LIMIT ? OFFSET ?`,
   `COUNT(*)` terpisah → meta.

**DoD:** curl semua kombinasi filter mengembalikan data + meta benar; sort field di luar whitelist
ditolak/di-fallback ke default (bukan error SQL); injection attempt aman.

### Fase 3 — PUT + soft DELETE (Task 1) + hide soft-deleted (Task 4)
1. `PUT /api/tasks/:id`: full update, validasi, tidak ada/sudah terhapus → `ErrNotFound` (404),
   duplikat title → 409.
2. `DELETE /api/tasks/:id`: `UPDATE ... SET deleted_at = NOW() WHERE id=? AND deleted_at IS NULL`
   → rows affected 0 = 404; sukses = 204.
3. Audit: `deleted_at IS NULL` ada di List, Update, Delete (dan semua query berikutnya).

**DoD:** curl PUT → 200 data berubah; PUT id tak ada → 404; DELETE → 204; task terhapus hilang dari
list; DELETE ulang task yang sama → 404.

### Fase 4 — Redis cache (Task 2)
1. `cache.go`: interface `TaskCache` + impl Redis.
   - Key kanonik dari SEMUA query param (field terurut): `tasks:list:keyword=fix&limit=10&page=1&sort=created_at:desc&status=todo`
     → dua request yang hanya beda urutan param → key sama → HIT.
   - `SET key <envelope json> EX 60`.
   - `InvalidateList`: `SCAN` match `tasks:list:*` + DEL batch.
   - Semua error cache di-log warning & diabaikan → Redis down tetap layani dari DB.
2. `service.go`: `List` = coba cache → miss → repo → set cache; `Create`/`Update`/`Delete` sukses →
   `InvalidateList()`.

**DoD:** curl GET dua kali → keduanya 200 (kedua dari cache — bukti: `redis-cli --scan --pattern 'tasks:list:*'`
+ log "cache hit"); setelah POST/PUT/DELETE → daftar key kosong lagi; matikan Redis → endpoint tetap 200.

### Fase 5 — Test (Task 5: update, search, cache invalidation)
Test persis yang diminta + sedikit pendukung — tidak mengejar coverage kosmetik:
1. `service_test.go` (mock repo + **miniredis**):
   - **Update**: update sukses mengubah field; id tak ada → `ErrNotFound`; title duplikat → `ErrDuplicateTitle`.
   - **Cache invalidation**: GET → repo terpanggil & key tersimpan; GET kedua → repo TIDAK terpanggil;
     Update → key hilang; GET ketiga → repo terpanggil lagi.
2. `repository_test.go` (**sqlmock**): **search** — kombinasi status/keyword/assignee menghasilkan
   WHERE benar; soft-deleted tidak masuk hasil; pagination LIMIT/OFFSET + COUNT benar.
3. (Pendukung, murah) satu httptest: POST duplikat → HTTP 409 dengan envelope — bukti langsung bullet Task 4.

**DoD:** `go test ./...` hijau, deterministik.

### Fase 6 — README + polish (Deliverables)
1. README: cara run (docker compose → migrate-up → run), env variables, tabel API + contoh curl,
   cara run test, catatan desain (cache strategy, soft delete, 409 mapping, out-of-scope/future work
   satu paragraf).
2. `gofmt` + `go vet` bersih (golangci opsional).
3. Commit rapi per fase (daftar di §8); `.env` tidak masuk git.

**DoD:** reviewer clone → ikuti README → semua jalan tanpa bertanya.

---

## 7. Pemetaan Bobot Nilai → di Fase Mana Dikerjakan

| Kriteria | Bobot | Fase |
|---|---|---|
| Go 35% | Fase 0–3, 6 (struktur berlapis, idiom, context, error handling) |
| Redis 15% | Fase 4 |
| SQL 10% | Fase 1–3 (skema, index, dynamic query, migration) |
| Testing 10% | Fase 5 |
| Code Quality 5% | Fase 6 |
| Documentation 5% | Fase 6 |

## 8. Commit Plan (conventional commits, per fase)

```
chore(backend): scaffold config, db/redis connection, docker-compose
feat(db): add tasks migration with soft delete and unique title
feat(api): create + list endpoints with centralized error envelope
fix(api): return 409 for duplicate title instead of 500
feat(api): list filtering, pagination, and whitelisted sorting
feat(api): PUT update and soft DELETE endpoints
feat(cache): cache task list 60s with query-param keys + invalidation on mutations
test(backend): update, search, and cache invalidation tests
docs(backend): README with setup, API contract, and tests
```

## 9. Checklist Requirement BE (review akhir, cocokkan per bullet)

- [ ] Filter: `status`, `keyword`, `assignee`, `page`, `limit`, `sort`
- [ ] `PUT /api/tasks/{id}` — 200/400/404/409
- [ ] Soft `DELETE /api/tasks/{id}` — 204, hilang dari list, DELETE ulang → 404
- [ ] Error konsisten via middleware terpusat (semua endpoint)
- [ ] Cache 60s, key memuat query param
- [ ] Invalidate setelah create/update/delete
- [ ] Duplicate title → 409
- [ ] Refresh list after update — invalidasi cache di BE (+ refetch di fase FE)
- [ ] Soft-deleted tidak pernah muncul
- [ ] Test: update, search, cache invalidation
- [ ] README + migration up/down + `.env.example`

---

## Appendix — Frontend (fase berikutnya, JANGAN dikerjakan sebelum BE DoD semua)

React Native + TS: `src/{api,components,screens,hooks,types}`, hook `useTasks` (loading/error/refetch,
AbortController), SearchInput (debounce ±400ms), StatusFilter, Pagination (ganti filter → reset page 1),
EditTaskModal, FlatList, skeleton/spinner, error banner.
Test: Jest + React Native Testing Library, minimal 1 (SearchInput debounce atau TaskList render).
Didiskusikan ulang saat BE selesai.

### Alur data EditTaskModal (keputusan desain — tanpa GET /api/tasks/:id)

```
GET /api/tasks → list state (TaskResponse memuat SEMUA field editable)
tap "Edit" di TaskItem → editingTask = task → modal render dengan initialValues={editingTask}
submit → PUT /api/tasks/:id → sukses → tutup modal → refetch() list
```

- Modal tidak fetch apa pun: object task sudah lengkap di memory, karena kontrak `TaskResponse` di list
  sengaja didesain memuat semua field yang bisa diedit. Ini keputusan kontrak API, bukan gaya kode.
- Trade-off (dicatat di README): risiko data basi hanya relevan untuk multi-user concurrency —
  di luar scope (tanpa auth/realtime). Setelah PUT selalu refetch → list selalu segar setelah edit.
- Opsi tengah kalau nanti dianggap perlu: tambahkan GET by id (~10 baris) + fetch saat modal dibuka.
  Ditunda karena tidak diminta bullet task manapun.
