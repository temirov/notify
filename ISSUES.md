# Issues

In this file the entries (issues) record newly discovered requests or changes, with their outcomes. No instructive content lives here. Read @NOTES.md for the process to follow when fixing issues.

Read @AGENTS.md, @ARCHITECTURE.md, @POLICY.md, @NOTES.md, @README.md and @ISSUES.md. Start working on open issues. Work autonomously and stack up PRs.

## Features  (102–199)

## Improvements (202–299)

## BugFixes (308–399)

## Maintenance (400–499)

- [x] [PG-400] Add profiles to @docker-compose.yml orchestration: dev for local build (using context and Dockerfiles) and docker for pulling all images from ghcr, including pinguin image. — docker-compose now exposes `dev` (local build) and `docker` (GHCR) profiles plus a regression test that enforces the layout.
- [ ] [PG-401] Only run GH docker-build.yml pipeline if go-test.yml workflow succeeds. Example 
```yml
on:
  workflow_run:
    workflows: ["tests"]
    types:
      - completed
```

## Planning
*do not work on these, not ready*
