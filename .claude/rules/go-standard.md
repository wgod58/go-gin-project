---
description: Go project layout and code standards for all Go files
globs: ["**/*.go"]
---

# Go Project Layout Standard

Follow the [golang-standards/project-layout](https://github.com/golang-standards/project-layout) conventions.

## Directory Structure

| Directory | Purpose |
|-----------|---------|
| `/cmd` | Main application entrypoints. Each sub-dir = one binary. Keep logic minimal — delegate to `/internal` or `/pkg`. |
| `/internal` | Private app code not importable by external projects. Go compiler enforces this. |
| `/pkg` | Library code safe for external use. Add deliberately — changes affect consumers. |
| `/api` | OpenAPI/Swagger specs, protobuf definitions, JSON schemas. |
| `/configs` | Config file templates and defaults. |
| `/scripts` | Build, install, analysis scripts. Keeps Makefile simple. |
| `/build` | CI configs and packaging (`/build/ci`, `/build/package`). |
| `/deployments` | Docker Compose, Kubernetes, Helm, Terraform. |
| `/test` | External test apps and test data (`/test/data`). |
| `/docs` | Design docs and user guides (supplements godoc). |
| `/tools` | Supporting tools; may import from `/pkg` and `/internal`. |
| `/examples` | Sample apps and usage demos. |
| `/vendor` | Vendored dependencies (`go mod vendor`). Optional with Go 1.13+ module proxy. |

**Never use `/src`** — this is a Java pattern and creates unnecessary nesting in Go.

## Code Conventions

- **Formatting**: All code must pass `gofmt`. Run before committing.
- **Linting**: Use `staticcheck` for code quality checks.
- **Modules**: Use Go Modules (`go.mod`/`go.sum`). Do not use GOPATH-based layout.
- **Package naming**: Lowercase, single word, no underscores or mixed caps (e.g., `userservice` not `UserService` or `user_service`).
- **Error handling**: Always handle errors explicitly. Do not use `_` to discard errors from functions that return them.
- **Interfaces**: Define interfaces in the consuming package, not the implementing package.
- **Internal packages**: Put application-specific code that shouldn't be shared in `/internal`. The Go compiler enforces this boundary.

## This Project's Layout

This project partially follows the standard. Current structure:

```
cmd/                  (if adding new binaries, place here)
config/               → maps to configs/ convention
controllers/          → belongs in internal/
services/             → belongs in internal/
models/               → belongs in internal/
middleware/           → belongs in internal/
interfaces/           → belongs in internal/ or pkg/
routes/               → belongs in internal/
grpc/                 → belongs in internal/
proto/                → belongs in api/
tests/                → belongs in test/
examples/             → standard /examples
docs/                 → standard /docs
```

When creating new packages, prefer the standard layout above over adding top-level directories.
