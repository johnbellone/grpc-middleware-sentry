# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Types of changes:
- `Added` for new features.
- `Changed` for changes in existing functionality.
- `Deprecated` for soon-to-be removed features.
- `Removed` for now removed features.
- `Fixed` for any bug fixes.
- `Security` in case of vulnerabilities.

## [0.4]
- Update repository configurations (vscode, github, dependabot, editorconfig)
- Prune dependencies (go mod tidy -compat=1.17)
- Bump several dependencies for security updates
- [@janwytze]: [PR#21] Fix parent transaction linking, add more details to transactions and update to newest Sentry SDK
- [@janwytze]: [PR#20] Add option to disable capturing the request body
- [@paulbrittain]: [PR#18] Add option to set operation name override.

## [0.3]
- Update to v0.20 version of Sentry SDK (and indirects).
- Update to v1.54.0 version of gRPC SDK (and indirects).
- Update to v1.4.0 version of gRPC Middleware (and indirects).

## [0.2]
- [@GTB3NW]: [PR#4] passes the span and context as intended by Sentry SDK.

## [0.1]
- [@slavaromanov]: [PR#3] exports functions and option types for client/server interceptors.

[Unreleased]: https://github.com/johnbellone/grpc-middleware-sentry/compare/v0.4.0...HEAD
[0.3]: https://github.com/johnbellone/grpc-middleware-sentry/tree/v0.3.0
[0.2]: https://github.com/johnbellone/grpc-middleware-sentry/tree/v0.2.0
[0.1]: https://github.com/johnbellone/grpc-middleware-sentry/tree/v0.1.0
[@slavaromanov]: https://github.com/slavaromanov
[@GTB3NW]: https://github.com/GTB3NW
[@paulbrittain]: https://github.com/paulbrittain
[@janwytze]: https://github.com/janwytze
[PR#21]: https://github.com/johnbellone/grpc-middleware-sentry/pull/21
[PR#20]: https://github.com/johnbellone/grpc-middleware-sentry/pull/20
[PR#18]: https://github.com/johnbellone/grpc-middleware-sentry/pull/18
[PR#4]: https://github.com/johnbellone/grpc-middleware-sentry/pull/4
[PR#3]: https://github.com/johnbellone/grpc-middleware-sentry/pull/3
