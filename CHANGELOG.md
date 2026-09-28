# CHANGELOG — Backend Task Management Assessment

> **Tujuan file ini:** jejak keputusan + progres pengerjaan BE. Cukup baca file ini untuk melanjutkan
> pekerjaan tanpa membaca ulang code / diskusi dari awal. Detail rasional: `DISCUSSION.md`.

---

## Aturan Kerja (WAJIB dibaca sebelum lanjut)

1. **Kerjakan PERSIS bullet task.** Dilarang menambah endpoint, file, library, atau fitur di luar
   daftar IN di bawah. Ini konteks "existing team" — **tidak mengubah / mengotak-atik code di luar
   scope task**.
2. Kebutuhan yang muncul di luar task → catat **1 baris** di README bagian "Future Work",
   **jangan** diimplementasi.
3. **Commit dilakukan oleh user sendiri** — pesan commit yang disiapkan ada di section bawah.
4. Dependensi: gin, go-sql-driver/mysql, go-redis (runtime) + miniredis, go-sqlmock (khusus test).
   Tidak ada dependensi lain.

---

## Status Saat Ini (per 2026-09-27)

- [x] Perencanaan & diskusi scope
- [x] **Implementasi Fase 0–6 — SELESAI** (`go build` / `go vet` / `gofmt` / `go test ./...` semuanya hijau,
      coverage package task: 65.4%)
- [x] `docker-compose.yml` **DIHAPUS** dari deliverable (final — analisa & alasannya di Catatan Implementasi #0;
      provisioning MySQL+Redis kini dijelaskan di README)
- [x] **Verifikasi runtime — SELESAI (2026-09-27, dieksekusi oleh Claude atas permintaan user)**:
      T1–T8 lulus semua terhadap server yang berjalan (bukti per kasus di section "Verifikasi Runtime").
      DB di-reset ke kondisi kosong setelahnya, siap untuk testing ulang oleh user. Satu temuan runtime
      sudah diperbaiki (Catatan Implementasi #9).
- Commit & push ke `main`: **dilakukan 2026-09-28** (bersama FE + docs; lihat "Pesan Commit").

---

## Keputusan Final (LOCKED — jangan diubah tanpa konfirmasi)

| # | Keputusan |
|---|---|
| 1 | **Scope IN** (semuanya dari bullet soal): GET list + filter (`status, keyword, assignee, page, limit, sort`), POST create (base "existing", prasyarat Task 2 & 4), PUT, soft DELETE, error envelope terpusat, Redis cache list 60s + invalidation, fix 409, hide soft-deleted, test update/search/cache-invalidation, README, migration up+down |
| 2 | **Scope OUT** (tidak dibuat): `GET /api/tasks/:id`, auth/tabel users, Dockerfile app, `docker-compose.yml` (**dihapus** — lihat Catatan #0), ORM, custom logger/recovery/CORS (pakai bawaan gin), singleflight, config library |
| 3 | `PUT` = **full update**: wajib `title` + `status`; opsional `description`, `assignee`, `due_date` |
| 4 | `sort` **satu param** `sort=field:asc\|desc`; whitelist: `title \| status \| assignee \| created_at \| updated_at`; default `created_at:desc`; di luar whitelist → fallback default |
| 5 | Duplicate title: `UNIQUE(title)` di DB + repository catch `1062` → `ErrDuplicateTitle` → 409 |
| 6 | `assignee` filter = **exact match**; `keyword` = `LIKE` pada `title` |
| 7 | Invalidasi cache = **SCAN prefix `tasks:list:*`** + DEL. Redis down → log warning, lanjut dari DB |
| 8 | Validasi list: `page ≥ 1` (default 1), `1 ≤ limit ≤ 100` (default 10), `status ∈ {todo, in_progress, done}` |
| 9 | Edit modal FE mengisi form dari list item (TaskResponse memuat semua field editable) → tidak butuh GET by id |
| 10 | Arsitektur: `handler → service → repository` satu arah via interface (`Service`, `Repository`, `TaskCache`) |

---

## Hasil Implementasi — Fase 0–6 (semua selesai)

### Fase 0 — Infra & skeleton — [x]
- [x] `go mod tidy` — temuan: `mongo-driver` & `quic-go` di go.mod ternyata **dependensi transitif
      resmi gin v1.12** (gin/binding→bson, gin→quic-go/http3), bukan sampah → dipertahankan
- [x] ~~`docker-compose.yml`~~ dibuat di fase ini, kemudian **DIHAPUS** (final — lihat Catatan Implementasi #0)
- [x] `internal/config/config.go` (stdlib env + default), `internal/database/{mysql,redis}.go` (pool + Ping fail-fast)
- [x] `cmd/api/main.go`: gin Logger+Recovery bawaan, `middleware.ErrorHandler()`, graceful shutdown
- [x] `.env.example`, `.gitignore`, Makefile (`run/test/tidy/lint/migrate-up/migrate-down`; migrate via `go run github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1` — CLI tidak perlu di-install)
- Verifikasi: `go build ./...` OK

### Fase 1 — Migration + base "existing" + fix 409 — [x]
- [x] `migrations/000001_create_tasks.{up,down}.sql` (soft delete + `UNIQUE uq_tasks_title` + index composite)
- [x] `internal/task/{model,dto,errors}.go`
- [x] `internal/middleware/error.go` — envelope terpusat: handler hanya `_ = c.Error(err)`,
      middleware memetakan `ErrValidation`→400 / `ErrNotFound`→404 / `ErrDuplicateTitle`→409 / lainnya→500
- [x] `repository.go`: `Create` (catch 1062→`ErrDuplicateTitle`) + `List`
- [x] `handler.go`: POST (201/409), GET (200)
- Verifikasi: unit test `TestCreateMapsDuplicateKeyToDomainError`, `TestCreateDuplicateTitleReturns409` (httptest, 409 + envelope) PASS

### Fase 2 — Filtering + pagination + sort — [x]
- [x] `normalizeListQuery` di service (default page/limit, clamp limit ≤100, default sort)
- [x] repo `List`: WHERE dinamis slice+placeholder `?`, `deleted_at IS NULL` selalu, whitelist sort
      (`orderClause`), `LIMIT ? OFFSET ?`, `COUNT(*)` terpisah → meta
- Verifikasi: `TestListAppliesAllFilters`, `TestListWithoutFiltersExcludesSoftDeleted`, `TestListSortWhitelist` (termasuk `evil; DROP TABLE tasks` → fallback) PASS

### Fase 3 — PUT + soft DELETE + hide soft-deleted — [x]
- [x] `PUT /api/tasks/:id` — full update; update → re-select row (response membawa timestamp segar);
      0 rows affected → `ErrNotFound`; duplikat → 409
- [x] `DELETE /api/tasks/:id` — `UPDATE ... SET deleted_at=NOW() WHERE id=? AND deleted_at IS NULL`;
      0 rows → 404; sukses → 204
- [x] `deleted_at IS NULL` terpasang di SEMUA query (List, Update, Delete, getByID) — dibuktikan regex query di test
- Verifikasi: `TestUpdateReturnsNotFoundWhenNoRowUpdated`, `TestUpdateReturnsFreshRow`, `TestSoftDelete` PASS

### Fase 4 — Redis cache — [x]
- [x] `cache.go`: interface `TaskCache` + impl Redis — key kanonik `tasks:list:assignee=...&keyword=...&limit=...&page=...&sort=...&status=...`
      (QueryEscape, urutan field tetap), `SET ... EX ttl` (default 60s), `InvalidateList` = SCAN + DEL batch,
      semua error cache di-log warning & diperlakukan sebagai miss (degrade)
- [x] `service.go`: List = cache → miss → repo → set cache; Create/Update/Delete sukses → `InvalidateList()`
- Verifikasi: `TestListCachesSecondRead` (repo hanya terpanggil 1×; key + TTL ada di miniredis),
      `TestListCacheKeyIncludesAllParams` PASS

### Fase 5 — Test: Update, Search, Cache invalidation — [x]
- [x] `service_test.go` (mock repo + miniredis): update sukses & invalidation, delete invalidation,
      update 404/409, create duplikat, normalize query
- [x] `repository_test.go` (sqlmock): search WHERE dinamis + argumen berurutan, soft-deleted tersaring,
      pagination (LIMIT/OFFSET + COUNT), whitelist sort, create/update/soft-delete
- [x] `handler_test.go` (httptest, package eksternal `task_test`): POST duplikat → 409 envelope
      (bukti Task 4)
- Verifikasi: **`go test ./...` hijau, 65.4% coverage package task** (scope test sengaja minimal sesuai Task 5)

### Fase 6 — README + polish — [x]
- [x] `README.md`: quickstart, env, API + contoh curl, error envelope, cache, test, design notes, Future Work
- [x] `gofmt -l .` bersih + `go vet ./...` bersih

---

## Verifikasi Runtime — SELESAI (2026-09-27, dieksekusi Claude atas permintaan user)

Dieksekusi terhadap server yang berjalan di mesin user (MySQL lokal user `adam`, Redis lokal).
Panduan perintahnya tetap di `backend/README.md` § "Testing Manual (curl)".

- [x] T1 Seed data — 3× POST → 201, id 1–3
- [x] T2 Duplicate title → **HTTP 409** + envelope `DUPLICATE_TITLE` (Task 4 ✓)
- [x] T3 Filter — status / keyword / assignee / kombinasi menghasilkan set yang tepat;
      status tak dikenal → 200 dengan `data: []`
- [x] T4 Pagination — p1l2 = 2 item + meta `{1,2,3,2}`; p2l2 = 1 item; `limit=0` → 10; `limit=500` → 100;
      `page=abc` → 400 `VALIDATION_ERROR` + fields
- [x] T5 Sort — asc/desc alfabet benar; string injection (`evil; DROP TABLE tasks`) → 200 normal,
      fallback `created_at:desc` (whitelist bekerja)
- [x] T6 PUT — 200 dengan timestamp segar (`updated_at` berubah); 404 id tak ada; 400 tanpa title;
      409 title dipakai task lain
- [x] T7 Soft delete — 204 → DELETE ulang 404 → task hilang dari list → PUT task terhapus 404
- [x] T8 Cache — key kanonik `tasks:list:...` terbentuk, TTL = 60; reorder param TIDAK menambah key
      baru (tetap HIT); invalidasi oleh create/update/delete → prefix `tasks:list:*` kosong

Catatan T8: sub-kasus "Redis di-stop → tetap 200" tidak dieksekusi (butuh sudo untuk stop service);
kebalikan perilaku yang sama sudah tercakup unit test `TestListDegradesWhenRedisIsDown`.

**T1–T8 lulus semua → syarat mulai Frontend (Task 3) terpenuhi.**

---

## Audit Scope & Testing Ulang — HASIL (2026-09-27, dieksekusi Claude; TANPA mengubah code)

Audit diminta user untuk memastikan (a) semua bullet task terpenuhi dan (b) tidak ada over-engineering.
Metode: pemeriksaan statis (file, dependensi, endpoint) + kualitas kode + testing runtime ulang dari
DB kosong. **Code tidak disentuh sama sekali selama audit ini.**

### A. Kepatuhan scope (anti over-engineering) — LULUS

- **Endpoint: tepat 4** (GET/POST `/tasks`, PUT/DELETE `/tasks/:id`) — dibuktikan dari registrasi router;
  tidak ada endpoint lain (tidak ada GET by id, tidak ada endpoint ekstra).
- **File: 24 file, semuanya ada fungsinya** — code domain `task/` (7 file), wiring (`main`, `config`,
  `database×2`, `middleware`), migration up/down, README, Makefile, `.env.example`, `.gitignore`;
  `.env` lokal di-gitignore. Tidak ada file "hiasan".
- **Dependensi langsung: tepat 5** — gin, go-sql-driver/mysql, go-redis (runtime) + miniredis,
  go-sqlmock (khusus unit test). Tidak ada dependensi lain; `mongo-driver`/`quic-go` di go.mod adalah
  transitif resmi gin v1.12.
- **Yang sengaja TIDAK dibuat** (dibandingkan terhadap soal): GET by id, auth/tabel users, ORM,
  CORS middleware, Dockerfile, singleflight, config library, pagination cursor, dsb. — tidak ada
  satu pun bullet soal yang memintanya.
- Dua hal yang layak dipertanyakan + justifikasinya: (1) graceful shutdown ±15 baris — praktik standar
  aplikasi Go, bagian kualitas Go 35%; (2) loader `.env` stdlib ±20 baris — supaya alur
  `cp .env.example .env` yang didokumentasikan README benar-benar bekerja, tanpa dependensi baru.
  Keduanya infrastruktur pelengkap, bukan fitur produk.
- **Kesimpulan A: tidak ada fitur/endpoint/library di luar bullet task.**

### B. Kualitas kode — LULUS

`go build` OK · `go vet` OK · `gofmt -l` bersih · **17 unit test lulus / 0 gagal** ·
coverage package `task` 66,8% (package lain = wiring murni; scope test sengaja minimal sesuai Task 5).

### C. Testing runtime ulang — 31/31 pernyataan perilaku BENAR

Dieksekusi terhadap server live (MySQL lokal `adam`, Redis lokal, DB mulai dari kosong).
Hasil per kelompok:

| Kelompok | Hasil |
|---|---|
| T1 Seed | 3× POST → 201; respons POST kini membawa `created_at` nyata (bukan zero-time) |
| T2 Duplicate | HTTP 409 + envelope `DUPLICATE_TITLE` |
| T3 Filter | status/keyword/assignee/kombinasi → set yang tepat; status tak dikenal → 200 kosong |
| T4 Pagination | meta `{page,limit,total_items,total_pages}` benar; clamp 0→10, 500→100; param kotor → 400 + fields |
| T5 Sort | asc/desc benar; string injection → fallback `created_at:desc`, tetap 200 |
| T6 PUT | 200 + data tersimpan; 404/400/409; body JSON rusak → 400 |
| T7 Soft delete | 204 → delete ulang 404 → hilang dari list → PUT task terhapus 404 |
| T8 Cache | key kanonik terbentuk, TTL 1–60; reorder param tanpa key baru; create/update/delete → prefix `tasks:list:*` kosong |

Satu baris "FAIL" di skrip audit terbukti **salah ekspektasi skrip, bukan bug produk**: kombinasi
`status=todo&assignee=budi` mengembalikan id `3|1` (bukan `1|3`) karena default sort `created_at:desc`
— diverifikasi ulang dengan `sort=title:asc` → `1|3` benar. Setelah T6 mengubah task 1 menjadi
`in_progress` (dan `assignee` menjadi NULL sesuai **PUT full-update**, Keputusan LOCKED #3), filter
kombinasi tetap mengembalikan set yang tepat — perilaku konsisten dari awal sampai akhir.

### Verdict Audit

**BE sesuai scope assessment 100% — tidak ada over-engineering, tidak ada fitur di luar soal,
semua verifikasi statis + unit + runtime lulus. Deliverable lengkap: git repo (commit menunggu user),
README, DB migration, unit tests.**

---

## Pesan Commit (untuk dijalankan USER — jangan commit dari sini)

File sudah saling terkait antar fase (mis. `main.go` memuat wiring dari beberapa fase), jadi dua opsi:

**Opsi A — satu commit utuh (paling praktis):**
```
feat(backend): task API with filtering, soft delete, error envelope, and redis cache
```

**Opsi B — per fase (riwayat lebih terbaca; urutkan seperti ini):**
```
chore(backend): scaffold config, mysql/redis connections, and makefile
feat(db): add tasks migration with soft delete and unique title
feat(api): create and list endpoints with centralized error envelope
fix(api): return 409 for duplicate title instead of 500
feat(api): list filtering, pagination, and whitelisted sorting
feat(api): implement PUT update and soft DELETE endpoints
feat(cache): cache task list for 60s with query-param keys and invalidate on mutations
test(backend): cover update, search, and cache invalidation
docs(backend): README with setup, API contract, and design notes
```

Catatan: `.env` sudah di-gitignore; jangan commit `.env`, cukup `.env.example`.

---

## Catatan Implementasi (deviasi & temuan — baca sebelum lanjut)

0. **docker-compose.yml DIHAPUS dari deliverable (final, 2026-09-27).** Analisa: assessment tidak
   pernah menyebut Docker — stack hanya mewajibkan MySQL & Redis **berjalan**, cara menyalakannya bebas.
   Mengikuti aturan "tidak membuat apa pun di luar task", file compose dikeluarkan dari repo; langkah
   provisioning MySQL+Redis (install native, atau container ad-hoc via `docker run`) didokumentasikan
   di README — README adalah deliverable yang diminta soal, dan menjelaskan cara menjalankan app adalah
   fungsinya. Perintah cache di guide testing memakai `redis-cli` dari host (kalau Redis di container:
   `docker exec <container> redis-cli ...`). Tetap tanpa Dockerfile untuk aplikasi (`go run`).
1. **Pola error middleware**: handler TIDAK memanggil formatter; handler melampirkan error dengan
   `_ = c.Error(err)` lalu `return` — `middleware.ErrorHandler()` (dipasang via `r.Use`) yang memformat
   setelah `c.Next()`. Alasan: menghindari import cycle (middleware mengimpor error domain `task`;
   handler ada di package `task`).
2. **`Update` di repository = UPDATE + re-select** agar response 200 membawa `created_at`/`updated_at`
   segar dari DB (bukan data buatan).
3. **mongo-driver & quic-go di go.mod itu sah** — dependensi transitif gin v1.12. Jangan dihapus manual;
   `go mod tidy` akan mengembalikannya.
4. **`TaskCache` bersifat nil-safe** di service (cache = optimasi, bukan dependency) — dipakai juga
   untuk kemudahan test.
5. Coverage 65.4% pada package task: sengaja — scope test persis Task 5 (update, search, cache
   invalidation) + 1 httptest 409, tanpa mengejar coverage kosmetik.
6. Migrasi dijalankan via `go run -tags mysql .../migrate/cmd/migrate@v4.18.1` di Makefile — reviewer
   tidak perlu install CLI apa pun. **Gotcha (sudah difix & diverifikasi 2026-09-27):** CLI migrate
   hanya mendaftarkan driver database lewat build tag — tanpa `-tags mysql` muncul error
   `"unknown driver mysql (forgotten import?)"`. Setelah ditambahkan `-tags mysql`, error berubah
   menjadi `connection refused` saat MySQL belum jalan (bukti driver sudah terdaftar).
7. **`.env` kini benar-benar dimuat** (fix 2026-09-27): `config.Load()` membaca file `.env` di working
   directory untuk variabel yang belum diset di environment (env asli tetap menang) — loader kecil
   stdlib tanpa dependensi baru. Sebelumnya `.env` hanya dokumentasi; app tetap jalan karena default
   config identik dengan `.env.example`, tapi `cp .env.example .env` jadi menyesatkan.
8. **Nama database default: `task_management`, user MySQL default: `adam`/`adamramadhans`** (menyesuaikan
   setup lokal user; konsisten di Makefile `DB_URL`, `.env.example`, default DSN di config, dan README).
9. **Fix respons POST (ditemukan saat verifikasi runtime, 2026-09-27)**: create sebelumnya mengembalikan
   `created_at/updated_at` zero-time (`0001-01-01`) karena INSERT tidak re-select — sekarang
   `Repository.Create` re-select seperti `Update`, jadi respons 201 membawa id + timestamp nyata dari DB.
   Mock & test repository/service ikut disesuaikan; ditambah test `TestListDegradesWhenRedisIsDown`.

---

## Checklist Requirement BE (dari bullet soal)

- [x] Filter: `status`, `keyword`, `assignee`, `page`, `limit`, `sort`
- [x] `PUT /api/tasks/{id}` — 200/400/404/409
- [x] Soft `DELETE /api/tasks/{id}` — 204, hilang dari list, DELETE ulang → 404
- [x] Error konsisten via middleware terpusat (semua endpoint)
- [x] Cache GET list 60s, key memuat query param
- [x] Invalidate setelah create/update/delete
- [x] Duplicate title → 409
- [x] Refresh list after update — invalidasi cache di BE (refetch di FE = fase frontend)
- [x] Soft-deleted tidak pernah muncul
- [x] Test: update, search, cache invalidation
- [x] README + migration up/down + `.env.example`
- [ ] Verifikasi runtime via Docker (checklist di atas) — **menunggu mesin dengan Docker**

---

## Langkah Berikutnya

1. ~~User: commit~~ — **selesai 2026-09-28** (verifikasi runtime & audit scope lulus; commit + push ke `main`).
2. ~~Frontend (Task 3)~~ — **selesai 2026-09-28**, semua fase FE-1–5 tereksekusi & terverifikasi (lihat `CHANGELOG-FE.md`).
