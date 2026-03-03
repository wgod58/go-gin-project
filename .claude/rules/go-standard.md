---
description: Go project layout and code standards for all Go files
globs: ["**/*.go"]
---

# Go Project Layout Standard

Reference: [golang-standards/project-layout](https://github.com/golang-standards/project-layout)

> This is **not an official standard** defined by the core Go dev team. It is a set of common historical and emerging project layout patterns in the Go ecosystem.

## Go Directories

### `/cmd`
Main applications for this project. Each subdirectory name should match the executable name (e.g., `/cmd/myapp`). Keep `main` functions minimal — they should only import and invoke code from `/internal` and `/pkg`. Don't put a lot of code here.

### `/internal`
**Private application and library code.** The Go compiler enforces this — packages inside `internal/` cannot be imported by code outside the parent of the `internal/` directory.

Sub-structure pattern for larger projects:
- `/internal/app/<appname>` — application-specific private code (services, handlers, middleware, routes)
- `/internal/pkg/<privlib>` — shared private libraries (models, repositories, infrastructure adapters)

You can have multiple `internal/` directories at any level of your project tree.

### `/pkg`
Library code safe for use by external applications. Other projects can import these. Use `/internal` instead if you don't want the code to be importable externally.

### `/vendor`
Application dependencies (`go mod vendor`). Optional with Go 1.13+ module proxy.

## Service Application Directories

### `/api`
OpenAPI/Swagger specs, JSON schema files, protocol definition files (e.g., `.proto` files).

## Common Application Directories

### `/configs`
Configuration file templates or default configs (confd, consul-template files).

### `/scripts`
Scripts for build, install, analysis operations. Keeps the root Makefile small and simple.

### `/build`
Packaging and CI/CD:
- `/build/package` — cloud/container/OS package configs
- `/build/ci` — CI configs (Travis, Circle, Drone, GitHub Actions)

### `/deployments`
IaaS, PaaS, container orchestration configs: docker-compose, Kubernetes/Helm, Terraform. Sometimes called `/deploy`.

### `/test`
Additional external test apps and test data. Go ignores directories/files beginning with `.` or `_`.
- `/test/data` or `/test/testdata` — test data files Go should ignore

## Other Directories

### `/docs`
Design and user documents (in addition to godoc-generated documentation).

### `/tools`
Supporting tools. Can import from `/pkg` and `/internal`.

### `/examples`
Examples for your applications and/or public libraries.

### `/third_party`
External helper tools, forked code, third-party utilities (e.g., Swagger UI).

### `/githooks`
Git hook scripts.

### `/assets`
Images, logos, and other media assets for the repository.

## Directories You Should NOT Have

### `/src`
**Never use this.** It's a Java pattern. It creates unnecessary nesting and does not belong in Go projects.

## The `internal/` Sub-Pattern

For projects using `internal/app` + `internal/pkg`:

```
internal/
  app/
    <appname>/    # or flat: service/, handler/, middleware/
  pkg/
    <privlib>/    # shared private libraries
```

Example for this project:
```
internal/
  app/
    service/      # business logic (UserService, AuthService, PaymentService)
    handler/      # HTTP handlers
    middleware/   # JWT auth middleware
  pkg/
    model/        # domain entities + repository/service interfaces
    repository/   # GORM models + DB implementations
    cache/        # Redis cache implementation
    stripe/       # Stripe client implementation
```

## Key Principles

- **Start simple**: For small projects, `main.go` + `go.mod` is enough. Add structure as you grow.
- **Use Go Modules**: `go.mod`/`go.sum` — don't use GOPATH-based layout.
- **`internal/` over `pkg/`**: Use `internal/` to enforce package privacy. The compiler enforces it.
- **Minimal `cmd/`**: Keep `main` functions thin — just wiring and startup.
- **`/api` for specs**: Put `.proto`, OpenAPI/Swagger, JSON schema files here.
- **`/test` for external tests**: Unit tests live alongside source files (`_test.go`). `/test` is for external integration tests and test data.

## This Project's Layout

```
go-gin-project/
├── main.go                    # wiring + server startup (HTTP :8080, gRPC :50051)
├── go.mod / go.sum
├── Makefile
├── docker-compose.yml
│
├── internal/
│   ├── app/
│   │   ├── service/           # UserService, AuthService, PaymentService
│   │   ├── handler/           # Gin HTTP handlers (user, auth, payment)
│   │   └── middleware/        # JWT auth middleware
│   └── pkg/
│       ├── model/             # User, Payment entities + repository/cache interfaces
│       ├── repository/        # GORM models + MySQL implementations
│       ├── cache/             # Redis cache implementation
│       └── stripe/            # Stripe client implementation
│
├── api/
│   └── proto/                 # .proto definitions + generated .pb.go
│
├── grpc/
│   ├── server/                # gRPC server
│   └── client/                # gRPC client + examples
│
├── config/                    # DB + env initialization (bootstrap, not business logic)
├── docs/                      # Swagger docs + design plans
├── examples/                  # Usage examples
└── test/
    └── mocks/                 # Mock implementations for testing
```

## Enforcement

Run before every commit:
```bash
gofmt -w .          # format
go vet ./...        # common mistakes
staticcheck ./...   # advanced analysis
go test ./... -v    # all tests pass
```
