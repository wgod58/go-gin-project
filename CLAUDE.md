# go-gin-project

Go REST API + gRPC server using Gin, GORM (MySQL), Redis, JWT auth, and Stripe payments.

Project-specific rules are in .claude/rules/ and are automatically loaded based on file path patterns.

## Commands

```bash
# Run tests
make test           # go test ./... -v
go test ./tests/... -v  # test specific package

# Docker services (MySQL + Redis)
make up             # start all services
make down           # stop all services
make init-db        # create goProject database
make mysql-cli      # connect to MySQL shell
make redis-cli      # connect to Redis shell

# Protobuf code generation
make proto-install  # install protoc-gen-go and protoc-gen-go-grpc
make proto          # regenerate from proto/*.proto

# Run app locally (requires .env and Docker services running)
go run main.go

# Run gRPC client example
make grpc-client
```

## Architecture

Two servers start concurrently in `main.go`:
- **HTTP** (port 8080) — Gin REST API with Swagger at `/swagger/*any`
- **gRPC** (port 50051, `$GRPC_PORT`) — user service via protobuf

```
main.go               Entry point, wires dependencies
config/config.go      DB init (GORM + AutoMigrate), env loading
models/               GORM models: User, Payment
services/             Business logic: user, auth, payment, redis cache
controllers/          Gin handlers: user, auth, payment
routes/routes.go      Route registration
middleware/auth.go    JWT auth middleware
interfaces/           Service interfaces (used for mocking)
grpc/server/          gRPC server implementation
grpc/client/          gRPC client + example
proto/                .proto definition + generated .pb.go files
tests/services/       Service unit tests (go-sqlmock)
tests/mocks/          Cache and Stripe mocks
docs/                 Swagger auto-generated docs
```

## Routes

```
POST /api/auth/login          public
POST /api/auth/admin-user     public (create user without auth)

# Protected (JWT required)
POST   /api/users/
GET    /api/users/:id
PUT    /api/users/:id
DELETE /api/users/:id

POST /api/payments/payment-intent
POST /api/payments/retrieve
```

## Environment Variables

Create a `.env` file (not committed):

```env
DB_USER=user
DB_PASSWORD=password
DB_HOST=localhost
DB_PORT=3306
DB_NAME=goProject

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=password

STRIPE_SECRET_KEY=sk_test_...
JWT_SECRET=your-secret

GRPC_PORT=50051   # optional, defaults to 50051
```

Docker Compose defaults: MySQL root password `password`, Redis password `password`.

## Testing

Tests use **go-sqlmock** for DB and custom mocks for Redis/Stripe in `tests/mocks/`.

```bash
go test ./tests/... -v -run TestUserService
```

Mock pattern: implement the interface from `interfaces/`, inject via constructor. See `tests/mocks/cache_mock.go`.

## Key Patterns

- **Dependency injection**: Services receive `*gorm.DB` and `interfaces.CacheInterface` via constructors.
- **Interface-based**: All external dependencies (DB, cache, Stripe) are behind interfaces for testability.
- **AutoMigrate**: `config.InitDB()` runs GORM AutoMigrate on startup for `User` and `Payment` models.
- **Swagger**: Run `swag init` after changing controller annotations to regenerate `docs/`.
