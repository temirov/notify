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
    - [ ] PN-100.A — Extend proto + domain models with `cancelled`/`errored` statuses, add list/update/cancel service methods, and cover new logic with table-driven tests.
    - [ ] PN-100.B — Introduce a Gin HTTP server that validates TAuth cookies (via `sessionvalidator`), exposes `/api/notifications` CRUD endpoints, and serves static assets alongside the gRPC process.
    - [ ] PN-100.C — Build the `/web` front-end skeleton (landing + dashboard pages, `constants.js`, `types.d.js`, `core/`, `ui/`, `app.js`) using Alpine factories and mpr-ui header/footer components.
    - [ ] PN-100.D — Implement the dashboard notifications table with inline edit/cancel flows, scheduled-time editor modal, status badges, and DOM-scoped Alpine events.
    - [ ] PN-100.E — Wire Google Identity Services + TAuth `auth-client.js` on both pages, manage auth state via Alpine store/BroadcastChannel, and enforce route guards/redirects.
    - [ ] PN-100.F — Add Playwright smoke tests that stub TAuth, verify landing-page login CTA, dashboard gating, table rendering, schedule edit, and cancel behavior.
    - [ ] PN-100.G — Expand `docker-compose.yaml` to run Pinguin + TAuth + static web host with shared JWT secret, and document the setup (env vars, npm test, new endpoints) in README/CHANGELOG.

## Improvements (200–299)

## BugFixes (300–399)

## Maintenance (400–499)

## Planning
*do not work on these, not ready*
