# Changelog

## Unreleased
- Added GoDoc coverage for all client-facing packages (client, attachments, grpcapi, grpcutil, logging) so integrators can rely on `go doc` to understand how to embed the SDK.
- Added a scheduling integration test backed by injectable senders to ensure emails queued with future timestamps are dispatched only after the background worker releases them.
- Extracted the scheduling/retry worker into `pkg/scheduler`, wired the server through a repository/dispatcher bridge, and added unit tests so other binaries can reuse the persistence-agnostic scheduler.
- Removed the `generate-secret` CLI command and `pkg/secret` helper in favor of documenting `openssl rand -base64 32` for token generation.
- Split the CLI and client-test utilities into standalone Go modules so `go run ./...` targets only the server binary.
- Documented all required environment variables in README/.env so the server starts without configuration surprises.
- Segregated server-only config, db, model, and service code under `internal/` while keeping shared clients in `pkg/`.
- Added a Cobra/Viper-based CLI for submitting immediate or scheduled notifications to the Pinguin gRPC service.
- Disabled SMS delivery when Twilio credentials are absent and emit a startup warning to document the configuration gap.
- Renamed the project to Pinguin, including module path, build targets, and user-facing documentation.
- Added comprehensive unit and integration tests across configuration, persistence, and gRPC layers.
- Added GitHub Actions workflow to enforce gofmt, go vet, and go test on pushes and pull requests.
- Added multi-stage Dockerfile and automated GHCR build workflow for the Pinguin gRPC server.
- Added optional `scheduled_time` to the gRPC Notification API and persisted model to support delayed dispatch.
- Updated retry worker to respect scheduled timestamps before attempting delivery.
- Introduced regression tests ensuring scheduled notifications remain queued until due.
- Migrated email delivery configuration to provider-agnostic SMTP naming, eliminating legacy third-party terminology from code and docs.
- Documented the SMTP delivery pipeline and added a unit test verifying the service wires the SMTP sender with configured credentials.
