# Issues

In this file the entries (issues) record newly discovered requests or changes, with their outcomes. No instructive content lives here. Read @NOTES.md for the process to follow when fixing issues.

## Features

- [x] [PN-05] Add a scheduler to the API and the implementation, alowing to schedule when the notifications are sent
  - Resolved: Introduced optional gRPC `scheduled_time`, persisted scheduling metadata, updated workers, and added scheduling regression tests.
- [x] [PN-08] Add a CLI client under new clients/cli folder. The CLI client shall be able to connect to the Pinguin Notification Service and submit/schedule notification delivery
  - Resolved: Added Cobra/Viper CLI with send command, injected gRPC client settings, scheduled-time parsing, and regression tests for request construction/error handling.
- [ ] [PN-16] Add a front-end.
   - have a config that defines admins, e.g. config.yml
   ```
   admins:
   - temirov@gmail.com
   - vadym@gmail.com
   ```
   - use this list to allow login through GAuss
   - display a table of all the received and delivered messages
   - have a filter for queued messages
   - Use footers and headers and stylign similar to Loopaware

## Improvements

- [x] [PN-07] Remove all and any mentioning of Sendgrid . Replace it with our own implementation of sending emails to the desination emails. Build a detailed plan of sending emails using SMTP protocol.
  - Resolved: Captured provider-agnostic SMTP delivery documentation, linked it from README, and added a wiring regression test for the in-process SMTP sender.
- [x] [PN-09] Disable SMS notifications and log the fact that the text notifications are disabled when TWILIO credentials are absent in the environemnt on the start
  - Resolved: Treated Twilio credentials as optional, logged the disabled state at startup, and prevented SMS dispatch/retries when configuration is incomplete.
- [x] [PN-11] Add generate-secret command to the CLI that generates sufficiently long secret string suited to be used as NOTIFICATION_AUTH_TOKEN. Keep the logic of the key generation in /pkg so that it could be later refactored into a shared package.
  - Resolved: Added pkg/secret crypto generator with length guards, wired `generate-secret` CLI command, and covered success/error flows with tests and documentation.
- [x] [PN-12] Remove generate-secret command and all associated files and document the usage of built in tools (openssl rand -base64 32) to get the strong secret key
  - Resolved: Deleted the CLI secret generator and related package, and updated README guidance to use `openssl rand -base64 32` for token creation.
- [x] [PN-18] Provide a Docker Compose example that persists the SQLite database on an external volume and documents how to run it end-to-end.
  - Resolved: Added `docker-compose.yaml` with a named Docker volume, shipped `.env.pinguin.example`, ensured the container image seeds `/var/lib/pinguin` with UID 65532 ownership for permission-safe mounts, documented compose usage in README, and verified go test/go vet run cleanly.

## BugFixes

- [x] [PN-06] Remove all and any mentioning, coding references or logic related to Sendgrid. The service is intended to be the required email integration with email providers, and not a middleman for other services.
  - Resolved: Renamed the email sender and configuration to provider-agnostic SMTP equivalents, updated env variables/tests, and refreshed documentation to eliminate SendGrid-specific language.
- [x] [PN-13] Fix the code so that there are no multiple entries:
  ```
  13:41:53 tyemirov@computercat:~/Development/pinguin [master] $ go run ./...
  go: pattern ./... matches multiple packages:
         github.com/temirov/pinguin/clients/cli
         github.com/temirov/pinguin/cmd/client_test
         github.com/temirov/pinguin/cmd/server
  ```
  - Resolved: Split the CLI and sample client into their own modules so root `go run ./...` targets only the server while dedicated commands still run locally.
- [x] [PN-15] Can not build client from the root folder:
```
14:55:27 tyemirov@computercat:~/Development/pinguin [master] $ go build -o bin/pinguin-cli clients/cli/main.go
clients/cli/main.go:7:2: no required module provides package github.com/spf13/viper; to add it:
        go get github.com/spf13/viper
clients/cli/main.go:8:2: no required module provides package github.com/temirov/pinguin/clients/cli/internal/command; to add it:
        go get github.com/temirov/pinguin/clients/cli/internal/command
clients/cli/main.go:9:2: no required module provides package github.com/temirov/pinguin/clients/cli/internal/config; to add it:
        go get github.com/temirov/pinguin/clients/cli/internal/config
```
  - Resolved: Added a Go workspace tying the server and CLI modules together, introduced a regression test ensuring `go build clients/cli/main.go` succeeds from the repository root, and verified fmt/vet/test suites.
- [x] [PN-17] Pinguin cant place a DB file when .env calls for a different path
```
16:15:52 tyemirov@computercat:~/Development/loopaware [improvement/LA-203-dashboard-footer] $ head -n3 .env.pinguin 
DATABASE_PATH=/var/lib/pinguin/pinguin.db
LOG_LEVEL=INFO
```
The error is
```
pinguin    | time=2025-10-30T23:15:44.638Z level=INFO msg="Starting gRPC Notification Server on :50051"
pinguin    | time=2025-10-30T23:15:44.638Z level=INFO msg="Initializing SQLite DB" path=/var/lib/pinguin/pinguin.db
pinguin    | time=2025-10-30T23:15:44.638Z level=ERROR msg="Failed to initialize DB" error="open sqlite failed: unable to open database file: no such file or directory"
pinguin exited with code 1 (restarting)
```
We shall be able to place the DB file on a docker image in order to preserve data continuity, and if we need to define the limits, we shall do so in README.md
  - Resolved: Added regression coverage for nested database paths, ensured `InitDB` creates parent directories before opening SQLite, and verified gofmt/go vet/go test ./... .


## Maintenance

- [x] [PN-01] Rename the project to Pinguin: repo, folder, all text references, all code reference. The project should be called Pinguin
  - Resolved: Renamed the module, regenerated gRPC assets, retitled documentation, and updated binaries/tests to use the Pinguin identity.
- [x] [PN-02] Cover the project with tests. The code must be fully tested
  - Resolved: Added configuration, persistence, service, and gRPC integration tests to exercise core behaviors with fakes and in-memory databases.
- [x] [PN-03] Add GitHub action for code testing
  - Resolved: Introduced a Go CI workflow executing gofmt checks, go vet, and go test for master pushes and pull requests.

  - Here is an example for inspiration:

  ```yaml
  name: Backend Tests

  on:
  push:
     branches:
        - master
     paths:
        - 'backend/**/*.go'
        - 'backend/go.mod'
        - 'backend/go.sum'
        - '.github/workflows/backend-tests.yml'
  pull_request:
     branches:
        - master
     paths:
        - 'backend/**/*.go'
        - 'backend/go.mod'
        - 'backend/go.sum'
        - '.github/workflows/backend-tests.yml'

  jobs:
  go-tests:
     name: Run Go unit tests
     runs-on: ubuntu-latest
     steps:
        - name: Checkout repository
        uses: actions/checkout@v4

        - name: Set up Go
        uses: actions/setup-go@v5
        with:
           go-version-file: backend/go.mod
           check-latest: true
           cache: true

        - name: Run go test ./...
        working-directory: backend
        run: go test ./...

  ```

- [x] [PN-04] Add GitHub action for building a docker container
  - Resolved: Added a multi-stage Dockerfile and CI workflow to publish the Pinguin server image to GHCR on master pushes.

  - Here is an example for inspiration

  ````yaml
  name: Build and Publish Backend Image

  on:
  push:
     branches:
        - master
     paths:
        - 'backend/Dockerfile'
        - 'backend/**/*.go'
        - 'backend/go.mod'
        - 'backend/go.sum'
        - '.github/workflows/backend-docker.yml'
  workflow_dispatch:

  jobs:
  build-and-push:
     name: Build and Push Image
     runs-on: ubuntu-latest
     permissions:
        contents: read
        packages: write

     steps:
        - name: Check out repository
        uses: actions/checkout@v4

        - name: Log in to GitHub Container Registry
        uses: docker/login-action@v3
        with:
           registry: ghcr.io
           username: ${{ github.actor }}
           password: ${{ secrets.GITHUB_TOKEN }}

        - name: Extract Docker metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
           images: ghcr.io/${{ github.repository_owner }}/gravity-backend
           tags: |
              type=raw,value=latest

        - name: Build and push Docker image
        uses: docker/build-push-action@v5
        with:
           context: ./backend
           file: ./backend/Dockerfile
           push: ${{ github.event_name != 'pull_request' }}
           tags: ${{ steps.meta.outputs.tags }}
           labels: ${{ steps.meta.outputs.labels }}
     ```
  ````

- [x] [PN-10] Refactor the code so that the code meant as internal structure of packages is under internal/ and the code emant to be shared with other programs is under pkg/
  - Resolved: Moved server-only config, db, model, and service packages under `internal/`, updated imports, and verified tests + vet across the tree.
- [x] [PN-14] The app failes to start with the message of missing variables but the readme doesnt mention them in the .env section. Add all missing variables to the README.md with the explanation of their meaning and suggested values
  ```
  pinguin    | time=2025-10-30T21:10:36.394Z level=ERROR msg="Configuration error" detail="configuration errors: missing environment variable OPERATION_TIMEOUT_SEC"
  pinguin    | time=2025-10-30T21:10:36.394Z level=ERROR msg="Configuration error" detail="missing environment variable CONNECTION_TIMEOUT_SEC"
  pinguin    | time=2025-10-30T21:10:36.394Z level=ERROR msg="Configuration error" detail="missing environment variable RETRY_INTERVAL_SEC"
  pinguin    | time=2025-10-30T21:10:36.394Z level=ERROR msg="Configuration error" detail="missing environment variable GRPC_AUTH_TOKEN"
  pinguin    | time=2025-10-30T21:10:36.394Z level=ERROR msg="Configuration error" detail="missing environment variable MAX_RETRIES"
  ```
  - Resolved: Updated README and the sample `.env` to enumerate every required variable (including GRPC and timeout settings) with recommended defaults.

## Planning (do not work on these, not ready)
