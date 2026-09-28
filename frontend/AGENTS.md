This is a **bare React Native + TypeScript** app (TaskManagement), scaffolded via
`@react-native-community/cli`. Scope is intentionally minimal — see
`../CHANGELOG-FE.md` (Keputusan Final FE) before adding anything.

## Commands

```bash
npm install            # install dependencies
npm start              # Metro bundler
npm run android        # build & run (Android emulator must be running)
npm run ios            # build & run (macOS + Xcode required)
npm test               # Jest, preset @react-native/jest-preset
npm run lint           # ESLint (@react-native/eslint-config)
npx tsc --noEmit       # typecheck
```

Run lint and typecheck before declaring any task done.

## Conventions

- The Go backend must be running (`make run` in `backend/`) for the app to load data.
- Base URL lives in ONE place: `src/api/client.ts` — Android emulator uses
  `http://10.0.2.2:8080` (localhost inside the emulator is the emulator itself),
  iOS simulator uses `http://localhost:8080`.
- API types in `src/types/task.ts` mirror `backend/internal/task/dto.go` — keep them
  in sync if the BE contract changes.
- Native folders `android/` and `ios/` exist; prefer JS/TS changes, only touch them
  when a dependency requires native linking.
- Tests use Jest + @testing-library/react-native (dev-only). No axios, no navigation
  library, no state library, no UI library — built-ins only.
