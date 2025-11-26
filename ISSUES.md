# Issues

In this file the entries (issues) record newly discovered requests or changes, with their outcomes. No instructive content lives here. Read @NOTES.md for the process to follow when fixing issues.

Read @AGENTS.md, @ARCHITECTURE.md, @POLICY.md, @NOTES.md, @README.md and @ISSUES.md. Start working on open issues. Work autonomously and stack up PRs.

## Features  (102–199)

## Improvements (202–299)

## BugFixes (308–399)

- [x] [PG-309] There is no more google sign in button in the header. There must have been an intgeration tests to verify it. — Restored `<mpr-login-button>` on landing/dashboard headers, re-seeded header attrs from tauth config, and reintroduced 14 Playwright scenarios that exercise Google/TAuth flows plus dashboard behaviors.

## Maintenance (400–499)

- [x] [PG-400] Add profiles to @docker-compose.yml orchestration: dev for local build (using context and Dockerfiles) and docker for pulling all images from ghcr, including pinguin image. — docker-compose now exposes `dev` (local build) and `docker` (GHCR) profiles plus a regression test that enforces the layout.
- [x] [PG-401] Only run GH docker-build.yml pipeline if go-test.yml workflow succeeds. Example 
```yml
on:
  workflow_run:
    workflows: ["tests"]
    types:
      - completed
```
 — docker-build workflow now listens to the Go Tests run completion and only pushes when the upstream job concludes successfully (manual dispatch retained for emergencies).
- [x] [PG-402] Remove the vendored `third_party` directory and rely on module dependencies so local copies of tauth/protobuf sources are no longer kept in the repo. — deleted the `third_party` tree, updated `go.work`, and confirmed builds/tests succeed without local tauth/protobuf copies.
- [x] [PG-403] Remove the standalone `cmd/client_test` module and relocate the integration test client into the `tests` package so all test tooling lives under the shared directory. — deleted the extra module, moved the helper CLI to `tests/clientcli`, updated README/Makefile/go.work, and confirmed tests still pass.
- [x] [PG-404] Relocate integration tests into `tests/integration`, rename the package to `integrationtest`, and ensure they target only the public surface so the suite reflects external usage. — tests now live under `tests/integration`, use the `integrationtest` package, and build tooling references the new path.
- [x] [PG-405] Remove the `third_party` directory entirely and rely strictly on upstream modules (TAuth validator, Google protos) without vendored copies. — verified the directory remains absent and added a regression test to fail CI if the vendored tree reappears.
- [x] [PG-406] Move `proto/pinguin.proto` under `pkg/` (e.g., `pkg/proto/`) so all shared API artifacts live with exported packages. — the proto now resides at `pkg/proto/pinguin.proto`, and README references the new location.
- [x] [PG-407] Relocate the CLI from `clients/cli` to `cmd/client`, fold it into the main module, and update documentation/build scripts accordingly. — CLI now lives under `cmd/client`, the standalone module was removed, and docs/tests reference the new path.
- [x] [PG-408] Remove `go.work`/`go.work.sum` once the extra modules are gone so the repository is managed solely via the root `go.mod`. — deleted the workspace files and confirmed `go test ./...` runs without them.
- [ ] [PG-410] Raise automated Go coverage to ≥95%. — Added regression tests for CLI config, logging helpers, generated proto/grpc bindings, the gRPC notification client, SMTP/Twilio senders, and retry dispatchers; repo-wide coverage climbed from 45% to 66.6%, but generated gRPC packages plus the server/service layers still drag the total below the target and require further investment.

## Planning
*do not work on these, not ready*
