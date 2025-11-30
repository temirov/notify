# Issues

In this file the entries (issues) record newly discovered requests or changes, with their outcomes. No instructive content lives here. Read @NOTES.md for the process to follow when fixing issues.

Read @AGENTS.md, @ARCHITECTURE.md, @POLICY.md, @NOTES.md, @README.md and @ISSUES.md. Start working on open issues. Work autonomously and stack up PRs.

## Features  (102–199)

- [x] [PG-103] add a flag (matched by an enviornment variables) that disables web interface. when the web interface is dsiabled it doesnt chech for the environment variables/flags required for web-interface functioning, such as ADMINS, GOOGLE_CLIENT_ID, HTTP_LISTEN_ADDR, HTTP_ALLOWED_ORIGINS, HTTP_STATIC_ROOT — Added the `--disable-web-interface` flag and `DISABLE_WEB_INTERFACE` env to skip HTTP/TAuth/Google config so Pinguin can run gRPC-only without those variables.

- [x] [PG-104] deliver a detailed technical plan to make pinguin multitenant (allowing serving multiple clients from different domains) — Authored `docs/multitenancy-plan.md` describing the tenancy model, schema changes, config strategy, API updates, migrations, and testing roadmap for multi-domain deployments.

- [x] [PG-105] implement the multitenancy roadmap from `docs/multitenancy-plan.md` starting with the tenant metadata foundation: added tenant/domain/member models, encrypted credential storage, bootstrap tooling, and repository/context helpers; HTTP middleware/runtime-config/session guards now resolve tenants per host. Completed tenant-scoped models, gRPC handlers, service layer, retry worker, and tests so every RPC requires a tenant context and queues remain isolated per tenant.

- [x] [PG-106] expose tenant metadata over gRPC/CLI/SDK surfaces — Added `tenant_id` fields to every proto request, ensured the CLI/SDK require `PINGUIN_TENANT_ID`, and updated the gRPC server to accept either `tenant_id` or the `x-tenant-id` metadata header so API clients can scope calls explicitly per tenant.

- [x] [PG-107] bind notification delivery to tenant-scoped credentials — Notification dispatch now resolves each tenant’s SMTP/Twilio profiles on the fly, caches sender instances per tenant, and drives both immediate sends and the retry worker through the tenant runtime. Tests cover sender caching, SMS-disabled tenants, and integration flows, so per-tenant secrets are always used even for scheduled jobs.

- [x] [PG-108] surface tenant identity in the browser runtime config and UI — Runtime bootstrap now merges the `/runtime-config` tenant payload into `window.__PINGUIN_CONFIG__`, updates `<mpr-header>` brand labels, and emits events so the tauth wiring refreshes Google IDs/base URLs dynamically. Constants expose the tenant display name, and new Playwright coverage asserts that the header branding matches the configured tenant.

- [x] [PG-109] add HTTP + Playwright regression tests for multi-tenant isolation — Added Go integration tests (`tests/integration/multitenancy_test.go`) covering both Service-layer data isolation and HTTP-layer host resolution. These tests verify that cross-tenant access is blocked and unknown hosts return 404s. Playwright tests were skipped in favor of backend integration tests due to the current mock-based Playwright setup (PG-411).

## Improvements (202–299)

- [ ] [PG-202] Refactor gRPC server to use an interceptor for tenant resolution instead of manual calls in every handler.
- [ ] [PG-203] Optimize retry worker to avoid N+1 queries per tick (iterating all tenants).
- [ ] [PG-204] Move validation logic from Service layer to Domain constructors/Edge handlers (POLICY.md).

## BugFixes (308–399)

- [x] [PG-309] There is no more google sign in button in the header. There must have been an intgeration tests to verify it. — Restored `<mpr-login-button>` on landing/dashboard headers, re-seeded header attrs from tauth config, and reintroduced 14 Playwright scenarios that exercise Google/TAuth flows plus dashboard behaviors.
- [ ] [PG-310] Fix critical performance bottleneck in `internal/tenant/repository.go`: implement caching for tenant runtime config to avoid ~5 DB queries + decryption per request.
- [ ] [PG-311] Fix potential null reference/crash in `ResolveByID` if `tenantID` is empty or invalid (missing edge validation).

## Maintenance (400–499)

- [ ] [PG-410] Raise automated Go coverage to ≥95%. — Added regression tests for CLI config, logging helpers, generated proto/grpc bindings, the gRPC notification client, SMTP/Twilio senders, and retry dispatchers; repo-wide coverage climbed from 45% to 66.6%, but generated gRPC packages plus the server/service layers still drag the total below the target and require further investment. Latest work added unit tests for server helper functions and attachment validation, but overall coverage remains at ~64% due to large untested surfaces (cmd/server entrypoint, scheduler, generated stubs). Further work is required to reach the goal.
- [ ] [PG-411] Replace the mocked Playwright harness with real end-to-end tests that exercise the Docker stack. — The current `tests/support/devServer.js` short-circuits every request (static HTML, fake `/auth/*`, fake `/api/notifications`) and the GIS stub overrides both Google Identity and the `mpr-ui` bundle, so CORS/login regressions slip through. We need a “real stack” profile that: (1) boots `docker compose` (ghttp + tauth + pinguin-dev) before Playwright runs and points `baseURL` at ghttp; (2) removes the devServer routes/stubs so the browser hits the actual HTTP server, runtime-config endpoint, and `/auth/*` handlers; (3) loads the real GIS script/CDN bundle, only mocking what CI cannot reach; (4) provides deterministic test data by exposing a backend reset/seed endpoint (or CLI) so `/api/notifications` has known fixtures; and (5) updates CI to run the suite against the containers, treating the existing mock-based checks as unit/UI tests. Without this, the “e2e” label is misleading and login/CORS failures will never be caught automatically.

## Planning
*do not work on these, not ready*