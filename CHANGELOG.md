# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- **The service now reads the serving knowledge base instead of the mining database.** The mining path is gone; there is no configuration that selects it. Endpoint routes and response shapes are unchanged
- **Search reads `component` instead of `projects ⋈ mines`.** The purl type was a `mines.purl_type` filter and is now the purl's own first segment, applied as `purl LIKE 'pkg:<type>/%'`. The popularity ordering is unchanged but reads `component_metrics`, which carries `git_stars`/`git_forks`/`git_created_at`/`first_version_date`/`versions`, one optional row per component
- **Versions read `component_version`** rather than `all_urls`: the release date comes from `component_version_status.indexed_date` and the licence from `component_version_license` through `license`. Whether a licence is SPDX-approved is no longer a stored flag - it is whether the identifier appears in `SPDX_license_data`. The package-level aggregate row (the empty version) is excluded
- **Status reads `component_status` and `component_version_status`**, keyed by `purl_id`, where mining kept the component status on `projects` and the version status on `all_urls`
- Component and version resolution no longer goes through `go-component-helper`, which resolves against `all_urls`, `mines` and `versions` - none of which exist in the serving schema. It is replaced by `pkg/resolver`
- An exact version now matches whether or not it carries a leading `v`, answering with the spelling the component publishes. The mining lookup did this through `SemverTogglePrefix`
- The queries use bound `?` placeholders through `Rebind`, so the same SQL runs on all three backends. The mining search wrote `$N` directly
- Binaries are built with `CGO_ENABLED=1`, and the runtime image moved to `debian:bookworm-slim` with `libstdc++6`. The DuckDB driver is a cgo binding, so a pure-Go build can no longer be produced

### Added
- **DuckDB support**, the intended production backend for the serving knowledge base. Set `DB_DRIVER=duckdb` and `DB_PATH` to the serving file, which is opened read-only
- `pkg/db`, opening the serving database for any of duckdb, sqlite or postgres
- `pkg/resolver`, the serving component resolution
- `pkg/models/db_version.go`. go-models' `DBVersionModel` selects a `created_at` column the serving `db_version` does not have, so the whole query failed and `/status` silently stopped reporting the knowledge base version. The local model reads what serving does have and derives the date from `db_release`, a `YY.MM.DD.HH.MM` build stamp
- `pkg/models/test_schema.go`, holding the serving schema the tests build their database from. The fixtures under `tests/` used to carry their own `CREATE TABLE` statements

### Fixed
- **A failing search pass no longer reads as "no components found".** `RunQueries` abandons the whole batch on the first error and the callers discarded that error, so a broken query was indistinguishable from a search that genuinely matched nothing. Two tests had encoded the resulting behaviour: searching npm for `react`, and searching for vendor `angular` + component `angular`, both reported nothing while the fixture carried matches

### Removed
- `pkg/models/all_urls.go`, replaced by `component_versions.go`. Nothing resolves through package URLs any more
- `preProcessQueryJob` and the `#ORDER` placeholder. The ranking clause is selected per purl type and written into the query directly

### Deployment
- **Breaking:** requires a serving knowledge base. Point `DB_PATH` at the `.duckdb` file and set `DB_DRIVER=duckdb`
- Coverage was checked against mining before the migration: `component` holds 29,950,449 rows against `projects`' 30,385,203 (98.6%); `component_status` and `component_version_status` cover every component and release; `component_metrics` covers 19.4M components, where mining had `git_stars` on only 3.97M. On a 40-component sample, 37 had identical version counts and 3 had one or two fewer
- Verified against the production serving file (55.7 GiB): search, versions, single and batch status, and the reported DB version
- `make build_arm` now cross-compiles with cgo and needs an `aarch64` toolchain (set `CC`); the amd64 build is unaffected

### Known issue (pre-existing, not from this change)
- `componentServer` embeds the `pb.ComponentsServer` **interface** rather than `pb.UnimplementedComponentsServer`, and that field is nil. Any call to a method the service does not implement dereferences it and **panics, killing the process**. `POST /v2/components/statistics` (`GetComponentStatistics`, added in papi v0.42.0 and never implemented here) does exactly that, so one unauthenticated request takes the server down. Embedding `pb.UnimplementedComponentsServer` instead would turn it into an `Unimplemented` error

## [0.11.0] - 2026-06-22
### Added
- `/health` liveness endpoint (GET) on the REST gateway
### Changed
- Upgraded `scanoss/go-grpc-helper` to `v0.16.0`

## [0.10.0] - 2026-04-30
### Added
- Enhanced component search to recover missing components
### Changed
- Updated dependencies to the latest versions

## [0.9.0] - 2026-04-21
### Changed
- Updated dependencies to the latest versions
- **BREAKING CHANGE:** Replaced `ErrorMessage`/`ErrorCode` with `InfoMessage`/`InfoCode` in `ComponentStatusResponse` for component and version status fields

## [0.8.0] - 2026-03-27
### Added
- Added `GetComponentStatus`/`GetComponentsStatus` services for getting single and multiple development life-cycle information
- Added support for custom status mapping for _retrieved_ and _registry-specific_ status
### Changed
- Using **go-component-helper** to get always the right component version based on user request


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


[0.11.0]: https://github.com/scanoss/components/compare/v0.10.0...v0.11.0
[0.10.0]: https://github.com/scanoss/components/compare/v0.9.0...v0.10.0
[0.9.0]: https://github.com/scanoss/components/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/scanoss/components/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/scanoss/components/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/scanoss/components/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/scanoss/components/compare/v0.4.0...v0.5.0
[0.0.1]: https://github.com/scanoss/components/compare/v0.0.0...v0.0.1
