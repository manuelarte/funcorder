# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.1.0] 2025-03-26

### Added

- Added linter to check for `NewXXX` or `MustXXX` functions to be after struct declaration but before struct methods.
- Add linter to check that struct's methods are after the struct delcaration and exported (public) methods are before not exported (private) ones.