# Task Management API — Backend

Backend untuk assessment Task Management: **Go (Gin) + MySQL 8 + Redis 7**.

## Scope yang diimplementasikan

- `GET /api/tasks` — filter `status`, `keyword`, `assignee` + pagination `page`/`limit` + `sort` (whitelist)
- `POST /api/tasks`, `PUT /api/tasks/:id` (full update), `DELETE /api/tasks/:id` (soft delete)
- Error response konsisten via middleware terpusat
- Cache Redis 60s untuk `GET /api/tasks`; key memuat semua query param; invalidasi setelah create/update/delete
- Duplicate title → `409`; soft-deleted task tidak pernah muncul di hasil manapun

## Prerequisites

- Go 1.27+
- MySQL 8 dan Redis 7 yang berjalan di `localhost` (install native, atau container — lihat Quickstart)
- Make

## Menjalankan Aplikasi

### Langkah 1 — Install MySQL & Redis (sekali saja)

Ubuntu/Debian:

```bash
sudo apt update
sudo apt install -y mysql-server redis-server
sudo systemctl enable --now mysql redis-server     # start + auto-start saat boot
```

Verifikasi: `systemctl is-active mysql redis-server` → `active` untuk keduanya.

### Langkah 2 — Siapkan database & user (sekali saja)

Kredensial di bawah cocok dengan default `.env.example` (`adam`/`adamramadhans`). Kalau mau pakai
kredensial lain, ubah `DB_DSN` di `.env` — jangan ubah keduanya sendiri-sendiri.

```bash
sudo mysql <<'SQL'
CREATE DATABASE IF NOT EXISTS task_management CHARACTER SET utf8mb4;
CREATE USER IF NOT EXISTS 'adam'@'localhost' IDENTIFIED BY 'adamramadhans';
ALTER USER 'adam'@'localhost' IDENTIFIED BY 'adamramadhans';
GRANT ALL PRIVILEGES ON task_management.* TO 'adam'@'localhost';
CREATE USER IF NOT EXISTS 'adam'@'%' IDENTIFIED BY 'adamramadhans';
ALTER USER 'adam'@'%' IDENTIFIED BY 'adamramadhans';
GRANT ALL PRIVILEGES ON task_management.* TO 'adam'@'%';
FLUSH PRIVILEGES;
SQL
```

(`ALTER USER` memaksa password-nya benar-benar `adamramadhans` — penting kalau user `adam` ternyata
sudah ada sebelumnya dengan password berbeda.)

Verifikasi — dua-duanya harus sukses sebelum lanjut:

```bash
mysql -uadam -padamramadhans -h 127.0.0.1 task_management -e "SELECT 1"   # → mencetak "1"
redis-cli ping                                                        # → PONG
```

### Langkah 3 — Konfigurasi & migrasi

```bash
cp .env.example .env     # boleh dilewati kalau .env sudah ada; sesuaikan DB_DSN bila perlu
make migrate-up          # membuat tabel tasks
```

### Langkah 4 — Jalankan server

```bash
make run                 # sama dengan: go run ./cmd/api
```

Log yang benar: `connected to mysql` → `connected to redis` → `listening on :8080`.
Biarkan terminal ini jalan, lalu cek cepat dari terminal lain:

```bash
curl -s http://localhost:8080/api/tasks
# → {"data":[],"meta":{"page":1,"limit":10,"total_items":0,"total_pages":0}}
```

Selanjutnya ikuti **Testing Manual (curl)** di bawah.

### Alternatif: MySQL & Redis via container (tanpa install native)

```bash
docker run -d --name tasks-mysql -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=rootpass -e MYSQL_DATABASE=task_management \
  -e MYSQL_USER=adam -e MYSQL_PASSWORD=adamramadhans mysql:8.0
docker run -d --name tasks-redis -p 6379:6379 redis:7-alpine
```

Setelah itu tetap lanjut dari Langkah 2 (buat user + grant) dan Langkah 3–4. Perintah `redis-cli`
di guide testing dijalankan dari host; kalau Redis-mu di container, bungkus:
`docker exec tasks-redis redis-cli ...`.

### Troubleshooting

| Error saat `make run` / `make migrate-up` | Penyebab | Perbaikan |
|---|---|---|
| `dial tcp 127.0.0.1:3306: connect: connection refused` | MySQL belum jalan | `sudo systemctl start mysql` |
| `dial tcp 127.0.0.1:6379: connect: connection refused` / `ping redis` gagal | Redis belum jalan | `sudo systemctl start redis-server`, cek `redis-cli ping` → PONG |
| `Error 1045 (28000): Access denied for user 'adam'@'localhost'` | User `adam` belum dibuat / password bukan `adamramadhans` | Jalankan ulang Langkah 2 (blok `sudo mysql` — `ALTER USER`-nya memaksa password) |
| `Error 1049: Unknown database 'task_management'` | Database belum dibuat | Bagian `CREATE DATABASE` di Langkah 2 |
| `unknown driver mysql (forgotten import?)` saat migrate | Flag build tag hilang | Sudah difix di Makefile (`go run -tags mysql ...`) — jangan hapus flag `-tags mysql` |

## Environment

| Variable            | Default                                                    | Keterangan                          |
|---------------------|------------------------------------------------------------|-------------------------------------|
| `APP_PORT`          | `8080`                                                     | Port HTTP server                    |
| `DB_DSN`            | `adam:adamramadhans@tcp(localhost:3306)/task_management?parseTime=true` | DSN MySQL — `parseTime=true` wajib |
| `REDIS_ADDR`        | `localhost:6379`                                           | Alamat Redis                        |
| `REDIS_PASSWORD`    | (kosong)                                                   | Password Redis                      |
| `CACHE_TTL_SECONDS` | `60`                                                       | TTL cache list                      |

## API

### GET /api/tasks

Query params:

| Param      | Aturan                                                                                       |
|------------|----------------------------------------------------------------------------------------------|
| `status`   | `todo` \| `in_progress` \| `done`                                                            |
| `keyword`  | `LIKE` pada `title`                                                                          |
| `assignee` | exact match                                                                                  |
| `page`     | ≥ 1, default 1                                                                               |
| `limit`    | 1–100, default 10                                                                            |
| `sort`     | `field:asc\|desc`; field whitelist: `title`, `status`, `assignee`, `created_at`, `updated_at`; default `created_at:desc`. Nilai di luar whitelist → fallback default (aman dari injection) |

```bash
curl "http://localhost:8080/api/tasks?status=todo&keyword=fix&page=1&limit=10&sort=created_at:desc"
```

```json
{
  "data": [
    { "id": 1, "title": "fix login", "description": null, "status": "todo",
      "assignee": "budi", "due_date": "2026-10-01",
      "created_at": "2026-09-27T08:00:00Z", "updated_at": "2026-09-27T08:00:00Z" }
  ],
  "meta": { "page": 1, "limit": 10, "total_items": 42, "total_pages": 5 }
}
```

### POST /api/tasks — create

```bash
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"fix login","status":"todo","assignee":"budi","due_date":"2026-10-01"}'
```

`201` dengan `{"data": {...}}` · `400` validasi · **`409` jika title duplikat**.

### PUT /api/tasks/:id — full update

Field wajib: `title`, `status`; field opsional boleh `null` (`description`, `assignee`, `due_date`).

```bash
curl -X PUT http://localhost:8080/api/tasks/1 \
  -H "Content-Type: application/json" \
  -d '{"title":"fix login v2","status":"in_progress"}'
```

`200` · `400` validasi · `404` tidak ada/sudah terhapus · `409` title dipakai task aktif lain.

### DELETE /api/tasks/:id — soft delete

```bash
curl -X DELETE http://localhost:8080/api/tasks/1   # → 204
curl -X DELETE http://localhost:8080/api/tasks/1   # → 404 (sudah terhapus)
```

Task terhapus tidak muncul lagi di `GET /api/tasks` (`deleted_at IS NULL` di semua query).

### Error envelope (semua endpoint)

```json
{ "error": { "code": "DUPLICATE_TITLE", "message": "task title already exists" } }
```

| Code               | HTTP | Keterangan                                    |
|--------------------|------|-----------------------------------------------|
| `VALIDATION_ERROR` | 400  | + field `fields` berisi pesan per input       |
| `NOT_FOUND`        | 404  |                                                |
| `DUPLICATE_TITLE`  | 409  |                                                |
| `INTERNAL_ERROR`   | 500  | detail error hanya ke log                      |

## Cache (Redis)

- Key memuat **semua** query parameter dalam urutan tetap, mis.
  `tasks:list:assignee=&keyword=fix&limit=10&page=1&sort=created_at%3Adesc&status=todo`
  → dua request yang hanya beda urutan parameter berbagi satu key (cache HIT).
- TTL 60 detik (`CACHE_TTL_SECONDS`); value = envelope lengkap `data` + `meta`.
- Setiap `POST`/`PUT`/`DELETE` yang sukses menghapus semua key `tasks:list:*`
  via `SCAN` (bukan `KEYS`, agar tidak memblokir Redis).
- Redis mati = degrade: error cache di-log sebagai warning, request tetap dilayani dari DB.

## Testing Manual (curl)

Panduan ini memetakan langsung ke bullet assessment (Task 1, 2, 4). Jalankan berurutan dan
cocokkan setiap hasil dengan tanda `→`.

### Persiapan

```bash
# Terminal 1 — prasyarat: selesaikan Langkah 1–2 di bagian "Menjalankan Aplikasi", lalu:
cp .env.example .env
make migrate-up
make run                      # biarkan jalan; log app tampil di sini

# Terminal 2
BASE=http://localhost:8080/api
```

### T1 — Seed data (3× POST → 201)

```bash
curl -s -X POST $BASE/tasks -H 'Content-Type: application/json' \
  -d '{"title":"fix login bug","status":"todo","assignee":"budi"}'
curl -s -X POST $BASE/tasks -H 'Content-Type: application/json' \
  -d '{"title":"write api docs","status":"in_progress","assignee":"sari"}'
curl -s -X POST $BASE/tasks -H 'Content-Type: application/json' \
  -d '{"title":"deploy staging","status":"todo","assignee":"budi"}'
```
→ tiga respons `{"data":{... "id":1 ...}}`, `"id":2`, `"id":3` (HTTP 201).

### T2 — Duplicate title → 409, bukan 500 (Task 4)

```bash
curl -s -i -X POST $BASE/tasks -H 'Content-Type: application/json' \
  -d '{"title":"fix login bug"}'
```
→ `HTTP/1.1 409 Conflict` dengan body:
`{"error":{"code":"DUPLICATE_TITLE","message":"task title already exists"}}`

### T3 — Filter: status, keyword, assignee (Task 1)

```bash
curl -s "$BASE/tasks?status=todo"                 # → hanya id 1 & 3
curl -s "$BASE/tasks?keyword=login"               # → hanya id 1 (LIKE di title)
curl -s "$BASE/tasks?assignee=budi"               # → id 1 & 3 (exact match)
curl -s "$BASE/tasks?status=todo&assignee=budi"   # → id 1 & 3 (kombinasi)
curl -s "$BASE/tasks?status=archived"             # → 200, "data":[] (tidak ada yg match)
```

### T4 — Pagination + meta (Task 1)

```bash
curl -s "$BASE/tasks?page=1&limit=2"   # → 2 item; meta {"page":1,"limit":2,"total_items":3,"total_pages":2}
curl -s "$BASE/tasks?page=2&limit=2"   # → 1 item
curl -s "$BASE/tasks?limit=0"          # → fallback ke default (meta.limit = 10)
curl -s "$BASE/tasks?limit=500"        # → clamp ke 100
curl -s -i "$BASE/tasks?page=abc"      # → 400 VALIDATION_ERROR + "fields"
```

### T5 — Sort (Task 1)

```bash
curl -s "$BASE/tasks?sort=title:asc"     # → deploy, fix, write (alfabet)
curl -s "$BASE/tasks?sort=title:desc"    # → write, fix, deploy
curl -s "$BASE/tasks?sort=created_at"    # → arah default desc (terbaru dulu)
curl -sG "$BASE/tasks" --data-urlencode 'sort=evil; DROP TABLE tasks'
# → tetap 200 normal, fallback ke created_at:desc (field di luar whitelist ditolak — anti-injection)
```

### T6 — PUT update (Task 1 + 4)

```bash
curl -s -X PUT $BASE/tasks/1 -H 'Content-Type: application/json' \
  -d '{"title":"fix login bug v2","status":"in_progress"}'
# → 200, data membawa title & status baru
curl -s "$BASE/tasks?keyword=v2"
# → perubahan langsung terlihat di list

curl -s -i -X PUT $BASE/tasks/999 -H 'Content-Type: application/json' -d '{"title":"x","status":"todo"}'
# → 404 {"error":{"code":"NOT_FOUND",...}}

curl -s -i -X PUT $BASE/tasks/1 -H 'Content-Type: application/json' -d '{"status":"todo"}'
# → 400 VALIDATION_ERROR (title wajib di full update)

curl -s -i -X PUT $BASE/tasks/1 -H 'Content-Type: application/json' -d '{"title":"write api docs","status":"todo"}'
# → 409 (title itu dipakai task 2)
```

### T7 — Soft delete + hide soft-deleted (Task 1 + 4)

```bash
curl -s -i -X DELETE $BASE/tasks/2    # → 204 No Content
curl -s -i -X DELETE $BASE/tasks/2    # → 404 (sudah terhapus)
curl -s "$BASE/tasks"                 # → task 2 tidak ada lagi di list
curl -s -i -X PUT $BASE/tasks/2 -H 'Content-Type: application/json' \
  -d '{"title":"zombie","status":"todo"}'
# → 404 (task terhapus tidak bisa di-update / dibangkitkan lagi)
```

### T8 — Cache Redis (Task 2)

Perintah `redis-cli` dijalankan dari host (Redis di `localhost:6379`). Kalau Redis-mu berjalan di
container, bungkus setiap perintahnya: `docker exec tasks-redis redis-cli ...`.

```bash
curl -s "$BASE/tasks?status=todo"     # request pertama → cache MISS (dari DB)
redis-cli --scan --pattern 'tasks:list:*'
# → muncul key, contoh:
#   tasks:list:assignee=&keyword=&limit=10&page=1&sort=created_at%3Adesc&status=todo

redis-cli TTL "tasks:list:assignee=&keyword=&limit=10&page=1&sort=created_at%3Adesc&status=todo"
# → angka 1–60 (detik tersisa; key expired otomatis max 60 dtk)

# Key kanonik: query sama dengan URUTAN PARAM BEDA → key yang sama (tetap HIT, tidak bikin key baru)
curl -s "$BASE/tasks?limit=10&page=1&status=todo&sort=created_at:desc"
redis-cli --scan --pattern 'tasks:list:*' | wc -l
# → tetap 1

# Invalidasi oleh CREATE
curl -s -X POST $BASE/tasks -H 'Content-Type: application/json' -d '{"title":"cache buster"}'
redis-cli --scan --pattern 'tasks:list:*'    # → kosong
# Invalidasi oleh UPDATE
curl -s "$BASE/tasks?status=todo"                                        # isi cache lagi
curl -s -X PUT $BASE/tasks/4 -H 'Content-Type: application/json' \
  -d '{"title":"cache buster 2","status":"done"}'
redis-cli --scan --pattern 'tasks:list:*'    # → kosong
# Invalidasi oleh DELETE
curl -s "$BASE/tasks?status=todo"                                        # isi cache lagi
curl -s -X DELETE $BASE/tasks/4
redis-cli --scan --pattern 'tasks:list:*'    # → kosong

# Degrade: Redis mati → endpoint tetap 200 dari DB (log app menampilkan [cache] warning).
# Hentikan Redis sesuai cara instalasimu (mis. systemctl stop redis-server / docker stop tasks-redis),
# pastikan `redis-cli ping` gagal, lalu:
curl -s -i "$BASE/tasks"              # → tetap HTTP/1.1 200 OK
# Nyalakan kembali Redis setelahnya.
```

### Reset data (opsional)

```bash
make migrate-down && make migrate-up   # kosongkan tabel dan mulai lagi
```

## Unit Tests (Task 5)

```bash
make test
```

Cakupan sesuai Task 5 — **update, search, cache invalidation**:

- `service_test.go` — update sukses/404/duplikat; list cache hit (read ke-2 tidak menyentuh repo);
  invalidasi oleh update & delete (mock repo + **miniredis**)
- `repository_test.go` — WHERE dinamis search, soft-deleted tersaring, pagination, whitelist sort
  (termasuk percobaan injection) (**sqlmock**)
- `handler_test.go` — POST duplikat → HTTP 409 dengan envelope (httptest)

`make lint` menjalankan `gofmt` + `go vet`.

## Design notes

- **Duplicate title → 409**: sumber kebenaran = `UNIQUE KEY uq_tasks_title`; repository menangkap
  MySQL error `1062` → `ErrDuplicateTitle` → middleware → 409. (Cek duplikat manual dengan `SELECT`
  itu race-prone.) Trade-off: dengan soft delete, judul yang sudah terhapus tetap "terpakai" —
  lihat Future Work.
- **Sort**: `ORDER BY` tidak bisa di-parameterize → field di-whitelist; di luar whitelist fallback
  ke `created_at:desc`.
- **Edit modal frontend** mengisi form langsung dari list item — `TaskResponse` sengaja memuat semua
  field yang bisa diedit, sehingga endpoint detail (`GET /api/tasks/:id`) tidak dibutuhkan.
- **Error handling**: handler hanya melampirkan error (`c.Error(err)`); satu middleware
  (`internal/middleware/error.go`) yang memformat envelope untuk semua endpoint.
- **Arsitektur**: `handler → service → repository` satu arah via interface (`Service`,
  `Repository`, `TaskCache`) → layer dapat di-mock untuk test.

## Structure

```
backend/
├── cmd/api/main.go              # entrypoint: config → koneksi → wiring → graceful shutdown
├── internal/
│   ├── config/config.go         # env + default (stdlib)
│   ├── database/{mysql,redis}.go
│   ├── middleware/error.go      # error domain → envelope HTTP (satu titik)
│   └── task/                    # model, dto, errors, repository, service, cache, handler (+ test)
├── migrations/000001_create_tasks.{up,down}.sql
├── Makefile · .env.example
```

## Future Work

Hal di luar scope assessment, sengaja tidak diimplementasi: endpoint detail `GET /api/tasks/:id`
(bila frontend nanti membutuhkan fetch-on-open), generated column `active_title` agar `UNIQUE(title)`
hanya berlaku untuk row aktif setelah soft delete, anti-cache-stampede (singleflight), auth +
tabel `users` (saat ini `assignee` string bebas), dan CI pipeline.
