# Issues

In this file the entries (issues) record newly discovered requests or changes, with their outcomes. No instructive content lives here. Read @NOTES.md for the process to follow when fixing issues.

## Features

- [ ] [PN-05] Add a scheduler to the API and the implementation, alowing to schedule when the notifications are sent

## Improvements


## BugFixes


## Maintenance

- [ ] [PN-01] Rename the project to Pinguin: repo, folder, all text references, all code reference. The project should be called Pinguin
- [ ] [PN-02] Cover the project with tests. The code must be fully tested
- [ ] [PN-03] Add GitHub action for code testing
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
- [ ] [PN-04] Add GitHub action for building a docker container
   - Here is an example for inspiration
   ```yaml
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



## Planning (do not work on these, not ready)


