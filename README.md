# Task Management — Fullstack Engineer Assessment

Aplikasi task management fullstack: **Go (Gin) + MySQL + Redis** untuk API,
**React Native + TypeScript** untuk aplikasi mobile. Repo ini berisi dua aplikasi
yang saling terpisah — keduanya punya README detail masing-masing.

```
├── backend/            # REST API — Go (Gin), MySQL 8, Redis 7
├── frontend/           # Aplikasi mobile — React Native 0.87 + TypeScript (bare RN)
├── CHANGELOG.md        # Jejak keputusan + hasil audit BE
├── CHANGELOG-FE.md     # Jejak keputusan + rencana eksekusi FE
└── DISCUSSION.md       # Rasional arsitektur (kenapa dipilih begini)
```

## Cakupan Assessment → Lokasi Implementasi

| Task | Soal | Implementasi | Status |
|---|---|---|---|
| 1 — Backend (40%) | filter `status/keyword/assignee/page/limit/sort`, `PUT /tasks/{id}`, soft `DELETE`, error konsisten | `backend/internal/task/` | ✅ Selesai + terverifikasi runtime (T1–T8) |
| 2 — Redis (15%) | cache `GET /api/tasks` 60s, key memuat query param, invalidasi setelah mutasi | `backend/internal/task/cache.go`, `service.go` | ✅ Selesai + terverifikasi runtime (T8) |
| 3 — Frontend (25%) | search input, status filter, pagination, edit modal, loading state | `frontend/src/{screens,components,hooks}` | ✅ Selesai (Fase FE-1–5) |
| 4 — Bug fixes (10%) | duplicate title → 409; refresh list setelah update; hide soft-deleted | 409: `repository.go` + `middleware/error.go`; hide: `deleted_at IS NULL` di semua query; refresh list: invalidasi cache BE + refetch FE setelah simpan modal | ✅ Selesai |
| 5 — Testing (10%) | BE: update, search, cache invalidation. FE: ≥1 component test | `backend/internal/task/*_test.go` (18 test) · `frontend/src/**/__tests__` (7 test, termasuk SearchInput debounce) | ✅ Semua hijau |

Detail perilaku tiap endpoint + cara mengujinya manual dengan curl:
lihat [`backend/README.md`](backend/README.md).

## Prasyarat

| Untuk | Perlu |
|---|---|
| Backend | Go 1.27+, MySQL 8, Redis 7, Make |
| Frontend | Node.js ≥ 22.11, JDK 17+, Android SDK (+ emulator), atau macOS + Xcode untuk iOS |

## Menjalankan (urutan: backend dulu, frontend butuh backend hidup)

### 1. Backend — API di `:8080`

Panduan lengkap (setup MySQL/Redis sekali jalan, troubleshooting, testing manual curl):
[`backend/README.md`](backend/README.md). Ringkasnya:

```bash
cd backend
cp .env.example .env          # sesuaikan DB_DSN bila perlu
make migrate-up               # buat tabel tasks
make run                      # → listening on :8080
```

Verifikasi cepat:

```bash
curl -s http://localhost:8080/api/tasks
# → {"data":[],"meta":{"page":1,"limit":10,"total_items":0,"total_pages":0}}
```

### 2. Frontend — app di emulator Android

```bash
cd frontend
npm install
npm start                     # Metro bundler — biarkan jalan
npm run android               # terminal lain; emulator harus sudah online
```

> **Gotcha klasik RN:** di Android emulator, `localhost` menunjuk emulator itu sendiri,
> bukan mesin host. Base URL sudah ditangani di satu tempat — `frontend/src/api/client.ts`
> (`http://10.0.2.2:8080` untuk Android emulator, `http://localhost:8080` untuk iOS).
> Perangkat fisik: ganti ke IP LAN mesin host.

## Testing & Lint

```bash
# Backend
cd backend && make test       # unit test (update, search, cache invalidation)
make lint                     # gofmt + go vet

# Frontend
cd frontend && npm test       # Jest + React Native Testing Library
npx tsc --noEmit              # typecheck
npm run lint                  # ESLint
```

## Ringkasan API

| Method | Path | Keterangan |
|---|---|---|
| `GET` | `/api/tasks` | list + filter + pagination + sort; cache Redis 60s |
| `POST` | `/api/tasks` | create; 409 jika title duplikat |
| `PUT` | `/api/tasks/:id` | full update (`title` + `status` wajib) |
| `DELETE` | `/api/tasks/:id` | soft delete (204; task hilang dari semua read) |

Error selalu satu bentuk: `{"error":{"code":"...","message":"..."}}` —
`VALIDATION_ERROR` 400 · `NOT_FOUND` 404 · `DUPLICATE_TITLE` 409 · `INTERNAL_ERROR` 500.

## Dokumentasi Lanjutan

- [`backend/README.md`](backend/README.md) — setup DB lengkap, env vars, kontrak API per endpoint,
  panduan testing manual T1–T8, design notes (kenapa 409 via UNIQUE key, whitelist sort, dll.)
- [`frontend/README.md`](frontend/README.md) — cara menjalankan app, base URL per platform, struktur kode
- [`CHANGELOG.md`](CHANGELOG.md) / [`CHANGELOG-FE.md`](CHANGELOG-FE.md) — keputusan desain terkunci
  dan alasannya, plus hasil audit scope (anti over-engineering) per sisi
- [`DISCUSSION.md`](DISCUSSION.md) — rasional arsitektur saat perencanaan
