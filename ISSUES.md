# Issues

In this file the entries (issues) record newly discovered requests or changes, with their outcomes. No instructive content lives here. Read @NOTES.md for the process to follow when fixing issues.

Read @AGENTS.md, @ARCHITECTURE.md, @POLICY.md, @NOTES.md, @README.md and @ISSUES.md. Start working on open issues. Work autonomously and stack up PRs.

## Features  (102–199)

- [x] [PG-103] add a flag (matched by an enviornment variables) that disables web interface. when the web interface is dsiabled it doesnt chech for the environment variables/flags required for web-interface functioning, such as ADMINS, GOOGLE_CLIENT_ID, HTTP_LISTEN_ADDR, HTTP_ALLOWED_ORIGINS, HTTP_STATIC_ROOT — Added the `--disable-web-interface` flag and `DISABLE_WEB_INTERFACE` env to skip HTTP/TAuth/Google config so Pinguin can run gRPC-only without those variables.

- [ ] [PG-104] deliver a detailed technical plan to make pinguin multitenant (allowing serving multiple clients from different domains)

## Improvements (202–299)

## BugFixes (308–399)

- [x] [PG-309] There is no more google sign in button in the header. There must have been an intgeration tests to verify it. — Restored `<mpr-login-button>` on landing/dashboard headers, re-seeded header attrs from tauth config, and reintroduced 14 Playwright scenarios that exercise Google/TAuth flows plus dashboard behaviors.

## Maintenance (400–499)

- [ ] [PG-410] Raise automated Go coverage to ≥95%. — Added regression tests for CLI config, logging helpers, generated proto/grpc bindings, the gRPC notification client, SMTP/Twilio senders, and retry dispatchers; repo-wide coverage climbed from 45% to 66.6%, but generated gRPC packages plus the server/service layers still drag the total below the target and require further investment.
- [ ] [PG-411] Replace the mocked Playwright harness with real end-to-end tests that exercise the Docker stack. — The current `tests/support/devServer.js` short-circuits every request (static HTML, fake `/auth/*`, fake `/api/notifications`) and the GIS stub overrides both Google Identity and the `mpr-ui` bundle, so CORS/login regressions slip through. We need a “real stack” profile that: (1) boots `docker compose` (ghttp + tauth + pinguin-dev) before Playwright runs and points `baseURL` at ghttp; (2) removes the devServer routes/stubs so the browser hits the actual HTTP server, runtime-config endpoint, and `/auth/*` handlers; (3) loads the real GIS script/CDN bundle, only mocking what CI cannot reach; (4) provides deterministic test data by exposing a backend reset/seed endpoint (or CLI) so `/api/notifications` has known fixtures; and (5) updates CI to run the suite against the containers, treating the existing mock-based checks as unit/UI tests. Without this, the “e2e” label is misleading and login/CORS failures will never be caught automatically.

## Planning
*do not work on these, not ready*
