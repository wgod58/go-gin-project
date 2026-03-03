---
description: Clean code standards combining Go idioms and general principles
globs: ["**/*.go"]
---

# Clean Code Standards

## Naming

- Package names: lowercase single word, no underscores (`userservice` → use `application` or `user` sub-package instead)
- Avoid redundant names: `user.UserService` → `user.Service`, `payment.PaymentRepo` → `payment.Repository`
- Exported types: PascalCase. Unexported: camelCase
- Acronyms: consistent casing — `userID` not `userId`, `httpServer` not `HTTPServer` in local vars
- Boolean vars/funcs: use `is`/`has`/`can` prefix: `isValid`, `hasPermission`

## Functions

- Single responsibility: one function does one thing
- Prefer ≤20 lines; if longer, extract helpers
- No boolean flag parameters: `process(user, true)` → two named functions
- Early returns over nested if-else
- Named return values only when they add clarity to complex returns

## Error Handling

- Always handle errors — never discard with `_` unless intentionally ignoring
- Wrap with context: `fmt.Errorf("create user: %w", err)` not `fmt.Errorf("error: %v", err)`
- No sentinel errors exported from packages — callers use `errors.Is` / `errors.As`
- Return errors, don't panic (except in `main` or truly unrecoverable situations)

## Packages

- Organize by responsibility, not by type (no `utils/`, `helpers/`, `common/`)
- Interfaces belong in the consuming package, not the implementing package
- Avoid circular dependencies — use the layer rule: `transport → application → domain ← infrastructure`
- Keep package surface small: prefer unexported types, export only what callers need

## Structs & Interfaces

- Small interfaces (1–3 methods): prefer `io.Reader` style over large interface blobs
- No god structs with 10+ fields doing multiple things
- Constructor functions (`NewXxx`) for all exported types that need initialization
- Don't export struct fields of internal types — use getters/setters or keep unexported

## Testing

- Table-driven tests for multiple cases: `tests := []struct{ name, input, want }{ ... }`
- Mock via interfaces — never mock concrete types
- No global state in tests — each test sets up its own dependencies
- Test file: `xxx_test.go` in same package (white-box) or `xxx_test` package (black-box)
- Each test must be independent and runnable in isolation

## Dependency Direction

```
transport (http/grpc handlers, middleware, routes)
    ↓ imports
application (services, use cases)
    ↓ imports
domain (entities, repository interfaces, service interfaces)
    ↑ implements
infrastructure (mysql repos, redis cache, stripe client)
```

- `domain` imports nothing from other internal layers
- `infrastructure` imports `domain` only (to implement interfaces)
- `application` imports `domain` only (uses interfaces, not concrete types)
- `transport` imports `application` only (calls service methods)
- `main.go` is the only file that imports all layers (for wiring)

## Tooling

Run before every commit:
- `gofmt -w .` — format all files
- `go vet ./...` — catch common mistakes
- `staticcheck ./...` — advanced static analysis
- `go test ./... -v` — all tests must pass
