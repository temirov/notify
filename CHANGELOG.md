# Changelog

## Unreleased
- Renamed the project to Pinguin, including module path, build targets, and user-facing documentation.
- Added comprehensive unit and integration tests across configuration, persistence, and gRPC layers.
- Added GitHub Actions workflow to enforce gofmt, go vet, and go test on pushes and pull requests.
- Added multi-stage Dockerfile and automated GHCR build workflow for the Pinguin gRPC server.
- Added optional `scheduled_time` to the gRPC Notification API and persisted model to support delayed dispatch.
- Updated retry worker to respect scheduled timestamps before attempting delivery.
- Introduced regression tests ensuring scheduled notifications remain queued until due.
- Migrated email delivery configuration to provider-agnostic SMTP naming, eliminating legacy third-party terminology from code and docs.
- Documented the SMTP delivery pipeline and added a unit test verifying the service wires the SMTP sender with configured credentials.
