# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- Upcoming changes...

## [0.7.0] - 2026-01-30
### Added
- Added database version info (`schema_version`, `created_at`) to `StatusResponse` across all component service endpoints
- Added server version to `StatusResponse`
- Log database version info on service startup
### Changed
- Moved server version from constructor parameter to `ServerConfig.App.Version`, configurable via `APP_VERSION` env var (defaults to embedded binary version)
- Log error when querying db version fails with an error other than `ErrTableNotFound`
- Updated `github.com/scanoss/go-models` to v0.3.0
- Updated `github.com/scanoss/go-grpc-helper` to v0.11.0

## [0.6.0] - 2025-09-18
### Added
- Added `name` field to component search and version response DTOs
### Changed
- Updated component DTOs to use `name` field instead of `component` field
- Upgraded `github.com/scanoss/papi` to v0.21.0
### Deprecated
- Deprecated `component` field in ComponentOutput and ComponentSearchOutput DTOs (use `name` instead)

## [0.5.0] - 2025-09-04
### Changed
- Removed `/api` prefix from REST endpoints
### updated
- Updated project dependencies to latest versions

## [0.0.1] - ?
### Added
- ?

[0.7.0]: https://github.com/scanoss/components/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/scanoss/components/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/scanoss/components/compare/v0.4.0...v0.5.0
[0.0.1]: https://github.com/scanoss/components/compare/v0.0.0...v0.0.1
