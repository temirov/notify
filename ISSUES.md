# Issues

In this file the entries (issues) record newly discovered requests or changes, with their outcomes. No instructive content lives here. Read @NOTES.md for the process to follow when fixing issues.

Read @AGENTS.md, @ARCHITECTURE.md, @POLICY.md, @NOTES.md, @README.md and @ISSUES.md. Start working on open issues. Work autonomously and stack up PRs.

## Features  (100–199)

- [ ] [PN-100] Add a front-end.
  Notes:
    - The front end is a stand alone web app. It uses Google Sign and TAuss backend for JWT generation and authentication
    - The front end has two pages: 
    1. The landing page, which carries marketing funbction and also has a login button
    2. The main page only accessible after the login:
      - The main page displays a table of all the received messages. it has a column for status: delivered/queued, a time of dlivery, a sender
      - The table allows editing of queued messages: changing the time of delivery or cancelling the delivery. The status column will have a cancelled and errored statuses.
      - The main page has a header and the footer from the mpr-ui library (@tools/mpr-ui)
    - Styling: 
      1. Use footers and headers and stylign from mpr-ui (@tools/mpr-ui)
      2. Support theme switch in the footer
      3. Have a landing page
    - Backend
    1. Integrate with TAuth service for the front end JWT verification (use the same secret). The TAuth service and its documentation is under @tools/TAuth. The landing page uses TAuth client to wrap the sign in requests
    2. The main page uses TAuth client to keep the user signed in
    3. Prepare docker compose orchestration
  Deliverable: perform the code analysis and prepare an implementation plan. AS we are integrating with two packages: mpr-ui and Tauth, be sure to consider the plan of integration. Deliver a plan expressed as small atomic tasks in @ISSUES.md
  Plan tasks:
    - [x] PN-100.A — Extend proto + domain models with `cancelled`/`errored` statuses, add list/update/cancel service methods, and cover new logic with table-driven tests. (proto + domain enums updated, new service APIs + retries covered by unit/integration tests)
    - [x] PN-100.B — Introduce a Gin HTTP server that validates TAuth cookies (via `sessionvalidator`), exposes `/api/notifications` CRUD endpoints, and serves static assets alongside the gRPC process. (Added config + Gin server with authenticated `/api/notifications` list/reschedule/cancel endpoints, session validation, static hosting, and docs/tests.)
    - [x] PN-100.C — Build the `/web` front-end skeleton (landing + dashboard pages, `constants.js`, `types.d.js`, `core/`, `ui/`, `app.js`) using Alpine factories and mpr-ui header/footer components. (Added landing + dashboard HTML, shared CSS, runtime config bootstrap, auth-aware Alpine components, and API client wiring.)
    - [x] PN-100.D — Implement the dashboard notifications table with inline edit/cancel flows, scheduled-time editor modal, status badges, and DOM-scoped Alpine events. (Added DOM-scoped events, toast center, validation, and table interactions wired to the HTTP API.)
    - [x] PN-100.E — Wire Google Identity Services + TAuth `auth-client.js` on both pages, manage auth state via Alpine store/BroadcastChannel, and enforce route guards/redirects. (Auth controller now listens to BroadcastChannel events, landing reflects GIS prep state, and dashboard redirects unauthenticated sessions immediately.)
    - [x] PN-100.F — Add Playwright smoke tests that stub TAuth, verify landing-page login CTA, dashboard gating, table rendering, schedule edit, and cancel behavior. (Playwright harness + mocked dev server now run via `npm test`.)
    - [x] PN-100.G — Expand `docker-compose.yaml` to run Pinguin + TAuth + static web host with shared JWT secret, and document the setup (env vars, npm test, new endpoints) in README/CHANGELOG. (Compose now runs the published `ghcr.io/tyemirov/tauth` image, shares signing keys via env files, and mounts `/web` so the UI + API are reachable on port 8080.)
  N.B. All code provided under tools/ is for information only and the tools/folder can never be referenced but the services shakll be referenced from their respective locations (github, CDN etc)

## Improvements (200–299)

- [x] [IM-200] Document docker orchestration quickstart.
  Notes:
    - README explains compose services but lacks a cohesive "start the stack" section.
    - Provide a dedicated quickstart outlining env file prep, `docker compose up`, and URLs (API vs UI) so newcomers can boot the full orchestration confidently.
    - Added a docker quickstart section (env copies, timed compose commands, port overview) plus changelog note.
- [ ] [IM-201] Only allow admins to log in through web interface. add ADMINS env var and make it a comma separated list, e.g. ADMINS=temirov@gmail.com,fivedoteight@gmail.com 

## BugFixes (300–399)

- [x] [BF-300] Dashboard API requests hit ghttp instead of the Pinguin HTTP server.
  Notes:
    - docker-compose now serves `/web` via ghttp on port 4173, but `web/js/constants.js` still uses `apiBaseUrl: '/api'`, so calls are sent to the static host.
    - Need to inject the correct API origin (e.g., `http://localhost:8080/api` or the service hostname inside docker) via runtime config/env. (Fixed by deriving the default from `window.location` and swapping 4173→8080 when detected.)
- [x] [BF-301] CORS/README instructions reference the wrong origin.
  Notes:
    - `.env.pinguin.example` keeps `HTTP_ALLOWED_ORIGINS=http://localhost:8080`, yet the UI now lives on `http://localhost:4173`, blocking compose-based testing.
    - README still tells users to browse `http://localhost:8080` for the landing page, leading to confusion and blank screens.
    - Updated the sample env + README docker-compose guidance to default to the ghttp UI origin (`http://localhost:4173`) and documented how to keep `HTTP_ALLOWED_ORIGINS` aligned so browsers can call the API.
- [x] [BF-302] Scheduled email integration test flakes.
  Notes:
    - `timeout -k 30s -s SIGKILL 30s go test ./integration -run ScheduledEmail -count 2` intermittently fails (`expected status sent, got queued`).
    - Logs show the worker context canceling before the scheduled notification flips to `sent`, so regression scenarios can slip past CI.
    - Need to stabilize the scheduler test by making notification timing deterministic (e.g., inject controllable clock/tick) or raising the worker wait to ensure the scheduled job executes before assertions.
    - Added a polling helper (`waitForNotificationStatus`) so the integration test waits for the persisted `sent` status instead of racing the worker, eliminating the flake.
- [x] [BF-303] Docker compose publishes ghttp on the wrong port.
  Notes:
    - `docker-compose.yaml` maps both `ghttp` and `pinguin` services to host port 8080, so the stack fails to start (`port is already allocated`).
    - README + `.env` expect the static bundle on `http://localhost:4173`, and CORS defaults now reference that origin.
    - Need to update compose to expose ghttp on 4173 (container 8080), ensure the HTTP server keeps port 8080, and document the change in CHANGELOG + sample env instructions if needed.
    - Updated `docker-compose.yaml` to publish ghttp on 4173, aligned `.env.tauth.example`/README guidance so TAuth CORS allows the same origin, and recorded the fix in the changelog.
- [x] [BF-304] CORS defaults allow credentialed requests from any origin.
  Notes:
    - When `HTTP_ALLOWED_ORIGINS` is empty, `buildCORS` sets `AllowAllOrigins=true` and `AllowCredentials=true`, so Gin echoes any `Origin` header while still sending cookies.
    - This effectively enables CSRF for all HTTP endpoints because any site can issue authenticated requests if a user has a TAuth session.
    - Need to either enforce an explicit allowlist or disable credentials when falling back to AllowAllOrigins (per requirement, disable credentials in the fallback).
    - Updated `buildCORS` to disable credentials for the fallback path and added unit tests verifying default vs allowlist behaviour.
- [x] [BF-305] Playwright fullyParallel mode races shared dev server state.
  Notes:
    - `playwright.config.ts` sets `fullyParallel: true`, so each test runs concurrently within a single file.
    - The mock dev server (`tests/support/devServer.js`) holds notifications in a process-wide array; tests rely on resetting it serially.
    - Parallel execution causes one test to cancel the only notification while another tries to reschedule it, intermittently disabling buttons and failing assertions.
    - Disable per-test parallelism (set `fullyParallel` to `false`) so smoke tests run sequentially until the dev server is isolated per worker.
    - Updated Playwright config, recorded the change in CHANGELOG, and reran the suite to confirm deterministic behaviour.
- [x] [BF-306] Static assets catch-all prevents HTTP server from starting.
  Notes:
    - `httpapi.NewServer` called `engine.StaticFS("/", …)` before registering `/api` routes, so Gin inserted a `/*filepath` catch-all and panicked because wildcards must be last.
    - Replaced the root `StaticFS` call with a `NoRoute` file server that serves files from `HTTP_STATIC_ROOT` only when no API route matches, ensuring the server boots cleanly.
    - Added a regression test that instantiates the server with a temp static root and verifies assets can be fetched without conflicting with `/api`.

## Maintenance (400–499)

## Planning
*do not work on these, not ready*
- [x] [BF-302] Inline config in `web/index.html` / `web/dashboard.html` overrides dynamic API base URL.
  Notes:
    - Even after deriving the default in `web/js/constants.js`, the HTML bootstrap sets `window.__PINGUIN_CONFIG__ = { apiBaseUrl: '/api', ... }`, so the browser still calls the static host when served from ghttp.
    - Remove the hard-coded `apiBaseUrl` (and ideally inject via env if needed) so the runtime detection takes effect.
