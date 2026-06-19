# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Extracted from `github.com/ocrosby/identity-platform-go/services/api-gateway` into a standalone repository under `jedi-knights/`. Module path is now `github.com/jedi-knights/api-gateway`. Per-line history is preserved via `git subtree split`; reviewers can `git blame` across the boundary.
- Initial scaffolding to match the `go-platform` / `go-logging` siblings: `Taskfile.yml`, `.golangci.yml` (US locale + `local-prefixes: github.com/jedi-knights/api-gateway`), `.github/workflows/ci.yml` (lint + test + build), `LICENSE` (MIT), `.gitignore`.

### Notes

The gateway depends on:

- `github.com/jedi-knights/go-platform` — `apperrors`, `container`, `httputil`
- `github.com/jedi-knights/go-logging/pkg/logging`

The container is wired through `platform.Container` with insertion-order `Bootstrap`, LIFO `OnClose`, and nil-interface-safe `Resolve`. See `CLAUDE.md` for the design contract.
