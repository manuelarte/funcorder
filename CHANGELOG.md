# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.5.0] 2025-05-09

### Removed

- Rollback `--fix` because of major bug [Issue #32](https://github.com/manuelarte/funcorder/issues/32).

## [v0.4.0] 2025-05-09

### Added

- Added `--fix` support.

## [v0.3.0] 2025-04-17

### Added

- Added `alphabetical` option.

## [v0.2.1] 2025-03-28

### Added

- Exporting setting constant names.

## [v0.2.0] 2025-03-28

### Added

- Added setting `constructor` to enable or disable Constructor check. 
- Added setting `struct-method` to enable or disable struct's method check.

## [v0.1.0] 2025-03-26

### Added

- Added linter to check for `NewXXX` or `MustXXX` functions to be after struct declaration but before struct methods.
- Add linter to check that struct's methods are after the struct declaration and exported (public) methods are before not exported (private) ones.