# Issues

In this file the entries (issues) record newly discovered requests or changes, with their outcomes. No instructive content lives here. Read @NOTES.md for the process to follow when fixing issues.

## Features

- [x] [PN-05] Add a scheduler to the API and the implementation, alowing to schedule when the notifications are sent
  - Resolved: Introduced optional gRPC `scheduled_time`, persisted scheduling metadata, updated workers, and added scheduling regression tests.
- [ ] [PN-08] Add a CLI client under new clients/cli folder. The CLI client shall be able to connect to the Pinguin Notification Service and submit/schedule notification delivery

## Improvements

- [x] [PN-07] Remove all and any mentioning of Sendgrid . Replace it with our own implementation of sending emails to the desination emails. Build a detailed plan of sending emails using SMTP protocol.
  - Resolved: Captured provider-agnostic SMTP delivery documentation, linked it from README, and added a wiring regression test for the in-process SMTP sender.
- [ ] [PN-09] Disable SMS notifications and log the fact that the text notifications are disabled when WILIO credentials are absent in the environemnt on the start

## BugFixes

- [x] [PN-06] Remove all and any mentioning, coding references or logic related to Sendgrid. The service is intended to be the required email integration with email providers, and not a middleman for other services.
  - Resolved: Renamed the email sender and configuration to provider-agnostic SMTP equivalents, updated env variables/tests, and refreshed documentation to eliminate SendGrid-specific language.

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

## Planning (do not work on these, not ready)
