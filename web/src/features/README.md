# Web feature modules

Product UI lives here, one folder per feature. Co-locate Vitest files
(`*.test.ts` / `*.test.tsx`) next to the code they cover.

| Folder | Notes |
|--------|--------|
| `search`, `clubs`, `matches`, `sharing` | Public journeys |
| `auth`, `preferences`, `push` | Account / alerts |
| `settings` | Account privacy UI (pairs with Go `internal/features/privacy`) |
| `admin`, `reports` | Maintainer panel and anonymous report form |

`web/src/pages/` is reserved for generic shells (`HomePage`, `NotFoundPage`,
`LoadErrorPage`). Do not add new product journeys there.

See `docs/adr/0002-package-layout.md`.
