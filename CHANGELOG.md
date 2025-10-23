# Changelog

## Unreleased
- Renamed the project to Pinguin, including module path, build targets, and user-facing documentation.
- Added comprehensive unit and integration tests across configuration, persistence, and gRPC layers.
- Added optional `scheduled_time` to the gRPC Notification API and persisted model to support delayed dispatch.
- Updated retry worker to respect scheduled timestamps before attempting delivery.
- Introduced regression tests ensuring scheduled notifications remain queued until due.
