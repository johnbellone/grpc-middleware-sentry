# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

Types of changes:
- `Added` for new features.
- `Changed` for changes in existing functionality.
- `Deprecated` for soon-to-be removed features.
- `Removed` for now removed features.
- `Fixed` for any bug fixes.
- `Security` in case of vulnerabilities.

## [Unreleased]

### Added

- Initial release of server/client interceptors for Sentry.
  - Report exceptions
  - Recover from panics
  - Export tags from context
  - Distributed tracing

[Unreleased]: https://github.com/johnbellone/grpc-middleware-sentry/compare/v1.0.0...HEAD
