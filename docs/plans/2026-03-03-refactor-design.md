# Refactor Design: Go Standard Layout + DDD Clean Architecture

**Date:** 2026-03-03
**Scope:** Add clean-code rule + restructure project to DDD layered architecture

---

## Goals

1. Add `.claude/rules/clean-code.md` covering Go idioms and general clean code principles.
2. Restructure the codebase to follow `golang-standards/project-layout` with a DDD flat-internal layering strategy.
3. Keep `main.go` at root (single binary, no `cmd/` needed).

---

## Clean-Code Rule

**File:** `.claude/rules/clean-code.md`
**Globs:** `["**/*.go"]`

Sections to cover:
- Naming (packages, exported/unexported, no redundancy)
- Functions (single responsibility, ≤20 lines, early returns, no bool flags)
- Error handling (always handle, wrap with `%w`, no sentinel errors in packages)
- Packages (by responsibility not type, interfaces in consumer)
- Structs & interfaces (small interfaces, no god structs)
- Testing (table-driven, mocks via interfaces, no global state)
- Dependency direction (`transport → application → domain ← infrastructure`)
- Tooling (`gofmt`, `staticcheck`, `golangci-lint`)

---

## Target Directory Structure

```
go-gin-project/
├── main.go                          # wires all layers, starts HTTP + gRPC
├── go.mod / go.sum
├── Makefile
├── docker-compose.yml
│
├── internal/
│   ├── domain/
│   │   ├── user.go                  # User entity + UserRepository interface
│   │   └── payment.go               # Payment entity + PaymentRepository interface
│   │
│   ├── application/
│   │   ├── user_service.go
│   │   ├── auth_service.go
│   │   └── payment_service.go
│   │
│   ├── infrastructure/
│   │   ├── mysql/
│   │   │   ├── user_repository.go   # implements domain.UserRepository
│   │   │   └── payment_repository.go
│   │   ├── redis/
│   │   │   └── cache.go             # implements CacheInterface
│   │   └── stripe/
│   │       └── client.go            # implements StripeService
│   │
│   └── transport/
│       ├── http/
│       │   ├── handler/             # user_handler.go, auth_handler.go, payment_handler.go
│       │   ├── middleware/          # auth.go
│       │   └── routes.go
│       └── grpc/
│           ├── server/
│           └── client/
│
├── api/
│   └── proto/                       # .proto + generated .pb.go
│
├── config/                          # DB + env initialization
├── docs/                            # Swagger docs + design plans
├── examples/
└── test/
    └── mocks/                       # cache_mock.go, stripe_mock.go
```

---

## Dependency Rules

```
transport  →  application  →  domain
                              ↑
                       infrastructure
```

- `domain`: no imports from other internal layers
- `application`: imports `domain` only
- `infrastructure`: imports `domain` (to implement interfaces)
- `transport`: imports `application` (calls services), never `infrastructure` directly
- `main.go`: the only file that imports all layers (wiring)

---

## File Migration Map

| Current | Target |
|---------|--------|
| `models/user.go` | `internal/domain/user.go` |
| `models/payment.go` | `internal/domain/payment.go` |
| `services/user_service.go` | `internal/application/user_service.go` |
| `services/auth_service.go` | `internal/application/auth_service.go` |
| `services/payment_service.go` | `internal/application/payment_service.go` |
| `services/redis_service.go` | `internal/infrastructure/redis/cache.go` |
| `controllers/user_controller.go` | `internal/transport/http/handler/user_handler.go` |
| `controllers/auth_controller.go` | `internal/transport/http/handler/auth_handler.go` |
| `controllers/payment_controller.go` | `internal/transport/http/handler/payment_handler.go` |
| `middleware/auth.go` | `internal/transport/http/middleware/auth.go` |
| `routes/routes.go` | `internal/transport/http/routes.go` |
| `grpc/server/` | `internal/transport/grpc/server/` |
| `grpc/client/` | `internal/transport/grpc/client/` |
| `proto/` | `api/proto/` |
| `tests/mocks/` | `test/mocks/` |
| `tests/services/` | `test/services/` |
| `interfaces/db_cache.go` | `internal/domain/` or `internal/application/` (interface belongs to consumer) |

DB repository interfaces currently implicit in services → extract to `internal/domain/`.

---

## Decisions

- GORM tags stay on domain entities (pragmatic — avoids dual model complexity for this project size)
- `config/` stays at top level (not inside `internal/`) — it's wiring/bootstrap code, not domain logic
- `DefaultStripeService` in `main.go` moves to `internal/infrastructure/stripe/client.go`
- Import path prefix stays `go-gin-project/` (no module rename needed)
