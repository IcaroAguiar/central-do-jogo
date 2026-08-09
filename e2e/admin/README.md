# Admin journey e2e (Playwright)

Skeleton for maintainer journeys (TEST-009): allowlist gate, source health,
correction/divergence, audit trail, and anonymous report queue.

Full suite delivery is **TASK-035**. Do not park admin Playwright specs under
`e2e/public/` — keep visitor smoke thin and admin coverage here (ADR 0002).

## Intended coverage (TASK-035)

- Maintainer session required for `/admin` and `/api/v1/admin/*`
- Non-maintainer receives 403
- Source health list and at-risk matches
- Confirm / correct / mark divergence with audit trail
- Anonymous report appears in maintainer queue without mutating match data

## Local run (when specs land)

Same stack prerequisites as `e2e/public/README.md`, plus a maintainer session
(stable test credentials or a seeded allowlist user). Then:

```bash
cd e2e/admin
pnpm exec playwright test --config=playwright.config.ts
```

Until TASK-035, this directory only holds the config/README placeholder so the
tree exists before the admin UI grows.
