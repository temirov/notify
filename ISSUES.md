# Issues

In this file the entries (issues) record newly discovered requests or changes, with their outcomes. No instructive content lives here. Read @NOTES.md for the process to follow when fixing issues.

Read @AGENTS.md, @ARCHITECTURE.md, @POLICY.md, PLANNING.md, @NOTES.md, @README.md and @ISSUES.md. Start working on open issues, prioritizing bug fixes. Work autonomously and stack up PRs.

## Features  (102–199)

## Improvements (202–299)

- [ ] [PG-202] Refactor gRPC server to use an interceptor for tenant resolution instead of manual calls in every handler.
- [ ] [PG-203] Optimize retry worker to avoid N+1 queries per tick (iterating all tenants).
- [ ] [PG-204] Move validation logic from Service layer to Domain constructors/Edge handlers (POLICY.md).

## BugFixes (308–399)

- [x] [PG-309] There is no more google sign in button in the header. There must have been an intgeration tests to verify it. — Restored `<mpr-login-button>` on landing/dashboard headers, re-seeded header attrs from tauth config, and reintroduced 14 Playwright scenarios that exercise Google/TAuth flows plus dashboard behaviors.
- [ ] [PG-310] Fix critical performance bottleneck in `internal/tenant/repository.go`: implement caching for tenant runtime config to avoid ~5 DB queries + decryption per request.
- [ ] [PG-311] Fix potential null reference/crash in `ResolveByID` if `tenantID` is empty or invalid (missing edge validation).

## Maintenance (400–499)



## Planning
*do not work on these, not ready*