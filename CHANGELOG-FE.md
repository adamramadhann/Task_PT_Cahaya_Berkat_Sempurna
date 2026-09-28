# CHANGELOG — Frontend Task Management Assessment (React Native + TypeScript)

> Jejak keputusan + rencana pengerjaan FE. Detail rasional BE: `CHANGELOG.md`, arsitektur: `DISCUSSION.md`.
> Prinsip sama dengan BE: **kerjakan PERSIS bullet task, tidak ada tambahan di luar soal.**

---

## Aturan Kerja (WAJIB dibaca sebelum eksekusi)

1. Scope FE = **Task 3 (5 bullet) + 1 component test (Task 5)**. Tidak ada layar, tombol, library,
   atau fitur lain di luar daftar IN di bawah.
2. Kebutuhan di luar task → catat 1 baris di README "Future Work", **jangan** diimplementasi.
3. Setiap fase selesai = DoD terbukti di emulator/perangkat + 1 commit (pesan disiapkan; commit
   dilakukan user).
4. Dependensi minimal: yang bawaan React Native + TypeScript. Tambahan HANYA dev-dependency test
   (`jest`, `jest-react-native` preset, `@testing-library/react-native`). **Tidak ada axios, tidak ada
   navigation, tidak ada UI library, tidak ada state library** — semua bisa pakai bawaan.

---

## Status Saat Ini (per 2026-09-28)

- [x] Perencanaan & penentuan scope FE
- [x] **Fase FE-0 selesai (kode)** — bare RN 0.87.1 + TS di `frontend/` (scaffold Expo lama dihapus,
  konfirmasi user → Opsi A bare RN, sesuai Keputusan #11); Jest preset bawaan + RNTL 14 (dev-only);
  `src/types/task.ts`, `src/api/client.ts`, `src/api/tasks.ts`, test placeholder `buildQuery`.
- [x] **Eksekusi Fase FE-1–5 SELESAI (2026-09-28)** — UI lengkap sesuai scope terkunci:
  `useTasks` (plain hooks + AbortController, race-safe), `TaskListScreen`, `TaskList`/`TaskItem`,
  `SearchInput` (debounce 400ms, fake-timer friendly), `StatusFilter` (chip + opsi "Semua",
  dipakai ulang di modal tanpa "Semua"), `Pagination` (dari `meta`, reset page saat filter/search
  berubah), `EditTaskModal` (prefill dari item tanpa GET by id, validasi title, spinner submit,
  pesan 409/400 tampil di dalam modal, sukses → `refetch()` = bug fix "refresh list after update").
- Verifikasi (2026-09-28): `npm test` **7/7 hijau** (component test SearchInput debounce dengan
  fake timers + smoke test app + test API layer), `npx tsc --noEmit` bersih, `npm run lint` bersih;
  app boot + load bundle dari Metro di emulator dengan backend live (11 task seed di DB).
- Catatan eksekusi RNTL 14: `render()` dan `fireEvent` **async** (wajib `await`) — pola test
  fake-timers: render dengan timer asli → `jest.useFakeTimers()` setelahnya → advance tanpa `act`.
- Commit & push ke `main`: dilakukan 2026-09-28.

---

## Keputusan Final FE (LOCKED — jangan diubah tanpa konfirmasi)

| # | Keputusan |
|---|---|
| 1 | **Scope IN**: Search input, Status filter, Pagination, Edit modal, Loading state, 1 component test — persis 6 bullet |
| 2 | **Scope OUT (JANGAN dibuat)**: form create task, tombol delete, layar kedua/navigation, Redux/Zustand/React Query, axios (pakai `fetch` bawaan), UI library (pakai `StyleSheet` bawaan), theme system, dark mode, i18n. Alasan semua: tidak ada di bullet Task 3. Catatan: BE punya DELETE & POST, tapi FE tidak diminta memakainya |
| 3 | **Single screen** — `App.tsx` langsung merender `TaskListScreen`; tidak perlu react-navigation |
| 4 | **State = plain hooks** (`useState`/`useEffect`/`useCallback`) di satu hook `useTasks`; tanpa library state |
| 5 | **Search debounce 400ms** di dalam `SearchInput`; `AbortController` membatalkan request basi saat user mengetik cepat (race condition) |
| 6 | **Ganti filter/search → reset page ke 1**; ganti halaman → fetch ulang |
| 7 | **Status picker = 3 chip tombol** (todo / in_progress / done) — tanpa paket picker tambahan |
| 8 | **Edit modal = `Modal` bawaan RN**, form terisi dari list item (tidak ada GET by id — sesuai keputusan BE); submit → `PUT` → sukses = `refetch()` list (bug fix "refresh list after update"); error 409 dari server ditampilkan di dalam modal |
| 9 | **Loading states**: spinner/skeleton saat fetch list, disable + spinner tombol saat submit, error banner + tombol retry |
| 10 | **API base URL di satu konstanta** (`src/api/client.ts`): iOS simulator `http://localhost:8080`, Android emulator `http://10.0.2.2:8080` — gotcha klasik RN |
| 11 | **Bare React Native + TypeScript** via `npx @react-native-community/cli init` (bukan Expo) — sesuai stack di soal; bisa diveto kalau mau Expo demi kemudahan review |
| 12 | **Test**: Jest + React Native Testing Library (dev-only). Minimal 1: `SearchInput` (debounce dengan fake timers) — boleh tambah `TaskList` render kalau waktu ada, tidak wajib |

---

## Kontrak dengan BE (yang dipakai FE — hanya ini)

| Dipakai | Endpoint |
|---|---|
| List | `GET /api/tasks?status&keyword&assignee&page&limit&sort` → `{data: Task[], meta: {page,limit,total_items,total_pages}}` |
| Edit | `PUT /api/tasks/:id` (full update: wajib `title`+`status`) → 200 · 400 · 404 · **409** envelope `{error:{code,message}}` |

Tipe `Task` (dari `TaskResponse` BE): `id, title, description, status, assignee, due_date, created_at, updated_at`.
`POST`/`DELETE` BE sengaja tidak dipanggil FE (tidak diminta Task 3).

---

## Struktur Folder FE (minimal)

```
frontend/
├── App.tsx                      # render TaskListScreen
├── src/
│   ├── api/
│   │   ├── client.ts            # fetch wrapper: base URL, JSON, error → throw bertipe
│   │   └── tasks.ts             # listTasks(query), updateTask(id, body) — typed
│   ├── components/
│   │   ├── SearchInput.tsx      # TextInput + debounce 400ms   (+ .test.tsx)
│   │   ├── StatusFilter.tsx     # 3 chip: todo/in_progress/done
│   │   ├── TaskItem.tsx         # 1 baris task, onPress → buka edit
│   │   ├── TaskList.tsx         # FlatList + empty state
│   │   ├── Pagination.tsx       # prev/next + "Halaman X dari Y"
│   │   └── EditTaskModal.tsx    # Modal + form + validasi + submit loading + pesan 409
│   ├── screens/
│   │   └── TaskListScreen.tsx   # komposisi semua komponen
│   ├── hooks/
│   │   └── useTasks.ts          # data/meta/loading/error + refetch + AbortController
│   └── types/
│       └── task.ts              # Task, ListMeta, Status
├── jest.config.js
├── package.json
└── tsconfig.json
```

---

## RENCANA EKSEKUSI — Fase FE-0–5 (berurutan; estimasi total ±6–7 jam)

### Fase FE-0 — Scaffold & infra (±1,5 j) — *prasyarat*
- [x] `npx @react-native-community/cli init TaskManagement --directory frontend` — bare RN 0.87.1 + TS
  (template kini menyertakan Jest + `react-test-renderer` + `__tests__/App.test.tsx` bawaan)
- [x] Setup Jest + React Native Testing Library — tambah `@testing-library/react-native` 14 (dev-deps);
  `jest.config.js` preset `@react-native/jest-preset` sudah dari template
- [x] `src/types/task.ts` + `src/api/client.ts` (fetch wrapper + base URL per platform, Keputusan #10)
- [x] `src/api/tasks.ts` — `listTasks`, `updateTask` typed; PUT di-unwrap dari envelope `{data: Task}`
- **DoD:** app default jalan di emulator/perangkat *(⏳ perlu user)*; `npm test` lulus (✅ 5/5:
  placeholder `buildQuery` + test template); tipe BE cocok (✅ dicek vs `dto.go` + `middleware/error.go`).
- Commit: `chore(frontend): scaffold react native + typescript, api client, and jest setup`

### Fase FE-1 — List dasar + loading/error (±1,5 j) — *Task 3: loading state* — [x]
- [x] `useTasks.ts`: state `{data, meta, loading, error}`, `refetch`, `AbortController`
- [x] `TaskListScreen` + `TaskList` (FlatList) + `TaskItem`
- [x] Loading: spinner saat fetch awal/ganti halaman; error banner + tombol retry; empty state
- **DoD:** dengan `make run` di BE → list tampil; matikan BE → muncul error + retry; refresh → data kembali.
- Commit: `feat(frontend): task list screen with loading, error, and empty states`

### Fase FE-2 — Search + status filter (±1 j) — *Task 3: search input, status filter* — [x]
- [x] `SearchInput` dengan debounce 400ms (fake-timer friendly)
- [x] `StatusFilter` 3 chip + opsi "semua"
- [x] Keduanya → ubah query `useTasks` → reset page 1; request lama dibatalkan AbortController
- **DoD:** ketik "login" → hasil terfilter (tanpa request per huruf — cek log BE); pilih status → terfilter; kombinasi keduanya benar.
- Commit: `feat(frontend): debounced search input and status filter`

### Fase FE-3 — Pagination (±1 j) — *Task 3: pagination* — [x]
- [x] `Pagination`: prev/next + "Halaman X dari Y" (dari `meta`), disable di ujung & saat loading
- [x] Ganti halaman → fetch `page=N`; ganti search/filter → kembali ke page 1
- **DoD:** butuh data > 1 halaman → seed minimal 3 task di BE (limit default 10 → seed ≥11 task, atau sementara set limit kecil via UI? tidak ada — pastikan seed ≥11) → navigasi halaman jalan, meta update.
- Commit: `feat(frontend): pagination controls driven by list meta`

### Fase FE-4 — Edit modal (±1,5 j) — *Task 3: edit modal + bagian loading; Task 4: refresh list* — [x]
- [x] `EditTaskModal`: Modal bawaan; form `title`, 3 chip status, `assignee`, `due_date` (opsional); pre-filled dari list item (tanpa GET by id)
- [x] Validasi client (title wajib); tombol save disable + spinner saat submit
- [x] Sukses → tutup modal → `refetch()` (bug fix "refresh list after update")
- [x] Error server tampil di modal: 409 `DUPLICATE_TITLE` → pesan jelas; 400 → pesan validasi
- **DoD:** edit title/status → list langsung berubah; coba title duplikat → pesan 409 di modal, modal tidak tertutup.
- Commit: `feat(frontend): edit task modal with validation and list refresh`

### Fase FE-5 — Component test + polish (±1 j) — *Task 5: minimal 1 test FE* — [x]
- [x] `SearchInput.test.tsx`: ketik → debounce (fake timers) → `onChangeText` terpanggil SEKALI setelah 400ms, bukan per huruf
- [x] (pengganti opsional TaskList test) `__tests__/App.test.tsx`: smoke test — app render sampai empty state dengan mock BE
- [x] README bagian FE: cara run (npm install / pod install / npx react-native run-android|ios), struktur, keputusan (base URL 10.0.2.2, full update PUT), Future Work (create/delete UI, React Query, navigasi)
- [x] Lint/format bersih (eslint + prettier bawaan template)
- **DoD:** `npm test` hijau; README FE lengkap; `git status` bersih dari file sampah.
- Commit: `test(frontend): SearchInput debounce component test` + `docs(frontend): README run instructions`

---

## Skenario Verifikasi Manual FE (jalankan setelah semua fase; BE harus jalan)

1. **List & loading** — buka app → data tampil; (opsional) throttle jaringan / matikan BE → error + retry → nyalakan BE → retry → data tampil.
2. **Search** — ketik kata kunci → setelah ±0,4 dtk hasil terfilter; kosongkan → kembali semua; ganti filter saat sedang mengetik → tidak ada hasil basi menempel.
3. **Status filter** — pilih `done` → hanya done; pilih `semua` → semua; kombinasi dengan search benar.
4. **Pagination** — seed ≥11 task → next/prev jalan, "Halaman X dari Y" benar, tombol disabled di ujung; ubah search → kembali ke halaman 1.
5. **Edit modal** — buka dari item → form terisi; ubah status → save → modal tutup + list ter-refresh dengan data baru; save dengan title kosong → ditolak client; save dengan title task lain → pesan 409 tampil di modal; submit → tombol disabled sampai selesai.

---

## Checklist Requirement FE (cocokkan per bullet saat review akhir)

- [x] Search input (debounce + race-safe)
- [x] Status filter
- [x] Pagination (dari meta BE, reset page saat filter berubah)
- [x] Edit modal (pre-filled dari list item, validasi, pesan 409)
- [x] Loading state (list, submit, error+retry, empty state)
- [x] Minimal 1 component test (SearchInput debounce)
- [x] README FE + lint bersih
- [x] Bug fix "refresh list after update" terasa di UI (refetch setelah PUT)

---

## Pesan Commit FE — SUDAH DIEKSEKUSI (2026-09-28)

Karena seluruh fase dieksekusi dalam satu sesi, riwayat digabung per area (bukan per fase):

```
feat(frontend): task list UI with debounced search, status filter, pagination, and edit modal
```

---

## Catatan untuk Sesi Berikutnya

- Semua fase FE selesai & terverifikasi (lihat Status Saat Ini) — fokus berikutnya: review HRD.
- Konfirmasi bare RN vs Expo (Keputusan #11) sudah selesai → **bare RN**; scaffold Expo lama dihapus.
- Pemetaan bobot: Frontend 20% (Fase FE-1–4), Testing 10% (bagian FE: Fase FE-5), Documentation 5% (README FE).
