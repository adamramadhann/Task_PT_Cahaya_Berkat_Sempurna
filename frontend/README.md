# Task Management — Frontend (React Native + TypeScript)

Aplikasi mobile untuk assessment Task Management: **bare React Native 0.87 +
TypeScript**, di-scaffold via `@react-native-community/cli` (bukan Expo).
Backend-nya adalah REST API Go di `../backend` — **app tidak menampilkan data
sebelum backend berjalan**.

## Status Implementasi

Sesuai rencana di [`../CHANGELOG-FE.md`](../CHANGELOG-FE.md) (scope terkunci) — **semua fase selesai**:

- ✅ **Fase FE-0** — scaffold bare RN + TS, Jest + React Native Testing Library,
  API layer (`src/api/`), tipe API (`src/types/task.ts` mencerminkan `dto.go` BE)
- ✅ **Fase FE-1–5** — task list + loading/error/empty state, SearchInput debounce
  400ms (race-safe via `AbortController`), StatusFilter chip, Pagination dari `meta`
  BE, EditTaskModal (prefill dari item, validasi title, spinner submit, pesan 409/400
  di dalam modal, refetch list setelah simpan), component test SearchInput (fake
  timers) + smoke test app. Verifikasi: `npm test` 7/7, `tsc --noEmit` bersih,
  `eslint` bersih, app boot di emulator dengan backend live.

## Prasyarat

- **Node.js ≥ 22.11** (`node --version`)
- **JDK 17+** dan **Android SDK** (ikuti [React Native — Set Up Your Environment](https://reactnative.dev/docs/set-up-your-environment))
- Emulator Android yang berjalan, atau perangkat fisik via `adb`
- iOS: hanya di macOS + Xcode (lihat bagian iOS di bawah)

## Menjalankan

**Backend harus jalan dulu** (lihat `../backend/README.md`): `make run` di `../backend`
→ API hidup di `:8080`.

```bash
npm install          # sekali, atau setelah package.json berubah
npm start            # Metro bundler — biarkan jalan di terminalnya
npm run android      # terminal lain — build + install + launch di emulator
```

Build pertama memakan waktu beberapa menit (Gradle). Setelah app terbuka, perubahan
kode JS/TS langsung ter-refresh (Fast Refresh) — tidak perlu build ulang.

### iOS (macOS saja)

```bash
bundle install               # sekali saja, untuk CocoaPods
bundle exec pod install      # sekali, atau setelah dependensi native berubah
npm run ios
```

### Base URL API — satu tempat

Semua request melewati `src/api/client.ts`. `localhost` **di dalam Android emulator
menunjuk emulator itu sendiri**, bukan mesin host — makanya:

| Target | Base URL |
|---|---|
| Android emulator | `http://10.0.2.2:8080` (alias emulator untuk loopback host) |
| iOS simulator | `http://localhost:8080` |
| Perangkat fisik | ganti manual ke IP LAN mesin host, mis. `http://192.168.1.10:8080` |

## Skrip

| Perintah | Fungsi |
|---|---|
| `npm start` | Metro bundler |
| `npm run android` / `npm run ios` | build + install + launch |
| `npm test` | Jest (preset `@react-native/jest-preset`) |
| `npm run lint` | ESLint (`@react-native/eslint-config`) |
| `npx tsc --noEmit` | typecheck |

Sebelum menyatakan selesai, jalankan: `npm run lint` + `npx tsc --noEmit` + `npm test`.

## Struktur Kode

```
frontend/
├── App.tsx                  # SafeAreaProvider + StatusBar → render TaskListScreen
├── src/
│   ├── api/
│   │   ├── client.ts        # fetch wrapper: BASE_URL, JSON in/out, error → ApiError bertipe
│   │   └── tasks.ts         # listTasks(query), updateTask(id, body) — endpoint yang dipakai FE
│   ├── components/
│   │   ├── SearchInput.tsx      # TextInput + debounce 400ms (+ __tests__)
│   │   ├── StatusFilter.tsx     # chip status — dipakai list (dengan "Semua") & modal (tanpa)
│   │   ├── TaskItem.tsx         # 1 baris task: title + badge status + assignee/due
│   │   ├── TaskList.tsx         # FlatList + spinner + empty state
│   │   ├── Pagination.tsx       # prev/next + "Halaman X dari Y" dari meta
│   │   └── EditTaskModal.tsx    # Modal bawaan: prefill, validasi, spinner, pesan 409/400
│   ├── screens/
│   │   └── TaskListScreen.tsx   # komposisi semua komponen di atas
│   ├── hooks/
│   │   └── useTasks.ts          # data/meta/loading/error + AbortController + refetch
│   ├── types/
│   │   └── task.ts              # cermin kontrak BE (`backend/internal/task/dto.go`) — jaga sinkron
│   └── api/__tests__/           # test unit API layer
├── android/ · ios/          # folder native — hindari menyentuh kecuali perlu linking native
└── jest.config.js · tsconfig.json
```

## Konvensi Proyek

- **Tanpa library tambahan di luar bawaan RN + TypeScript** — tidak ada axios
  (pakai `fetch`), navigation library, state library, atau UI library. Tambahan
  dev-dependency hanya untuk test (Jest, RNTL).
- Kontrak API yang dipakai hanya dua: `GET /api/tasks` dan `PUT /api/tasks/:id`.
  Endpoint `POST`/`DELETE` BE ada tapi sengaja tidak dipanggil FE (tidak diminta Task 3).
- Kalau kontrak BE berubah, sinkronkan `src/types/task.ts` dengan
  `backend/internal/task/dto.go`.
- Test komponen memakai **React Native Testing Library** — `SearchInput` (debounce,
  fake timers) di `src/components/__tests__/` + smoke test app di `__tests__/`.

## Troubleshooting

| Gejala | Penyebab | Perbaikan |
|---|---|---|
| `Tidak bisa menghubungi server di http://10.0.2.2:8080` di app | backend tidak jalan | `make run` di `../backend`, cek `curl localhost:8080/api/tasks` |
| `npm run android` gagal: *No connected devices* | emulator belum online | nyalakan emulator dulu (`avdmanager list avd` → `emulator -avd <nama>`), lalu ulangi |
| Build Gradle gagal soal SDK | platform/NDK tidak lengkap | buka `android/` di Android Studio → Sync + install yang diminta |
| App putih/blank setelah launch | Metro tidak terjangkau | pastikan `npm start` jalan; di emulator: `adb reverse tcp:8080 tcp:8080` tidak wajib karena base URL memakai `10.0.2.2` |
