# Go Payment Service

A REST API + gRPC service built with Go, integrating Stripe for payment processing, MySQL for data persistence, and Redis for caching. Structured following [golang-standards/project-layout](https://github.com/golang-standards/project-layout) with DDD layered architecture.

## Tech Stack

- **Go 1.26**
- **Gin** — HTTP web framework
- **GORM** — ORM for MySQL
- **Redis** — Caching layer
- **Stripe** — Payment processing
- **MySQL** — Primary database
- **gRPC + Protobuf** — RPC service
- **JWT** — Authentication
- **Docker** — Containerization
- **Swagger** — API documentation
- **go-sqlmock + testify** — Testing

## Project Structure

```
go-gin-project/
├── main.go                    # wires all layers, starts HTTP :8080 + gRPC :50051
├── config/                    # DB init and env loading
│
├── internal/
│   ├── app/
│   │   ├── service/           # business logic (UserService, AuthService, PaymentService)
│   │   ├── handler/           # Gin HTTP handlers (user, auth, payment)
│   │   ├── middleware/        # JWT auth middleware
│   │   └── routes.go          # route registration
│   └── pkg/
│       ├── model/             # domain entities + repository/service interfaces
│       ├── repository/        # GORM models + MySQL implementations
│       ├── cache/             # Redis cache implementation
│       └── stripe/            # Stripe client implementation
│
├── api/proto/                 # Protobuf definitions + generated Go code
├── grpc/                      # gRPC server + client
├── test/
│   ├── mocks/                 # Cache and Stripe mocks
│   └── services/              # Service unit tests
└── docs/                      # Swagger docs + design plans
```

## Architecture

Dependency flow: `handler → service → model interfaces ← repository/cache/stripe`

- **`internal/pkg/model`** — pure domain entities (`User`, `Payment`) and interfaces (`UserRepository`, `CacheService`, `StripeService`). No framework dependencies.
- **`internal/pkg/repository`** — GORM models with DB schema tags, MySQL implementations of domain interfaces.
- **`internal/pkg/cache`** — Redis implementation of `CacheService`.
- **`internal/pkg/stripe`** — Stripe SDK implementation of `StripeService`.
- **`internal/app/service`** — business logic, depends only on `model` interfaces.
- **`internal/app/handler`** — Gin HTTP handlers, depends only on services.
- **`main.go`** — the only file that wires all layers together.

## API Routes

```
POST /api/auth/login            # public — returns JWT
POST /api/auth/admin-user       # public — create user

# JWT required
POST   /api/users/
GET    /api/users/:id
PUT    /api/users/:id
DELETE /api/users/:id

POST /api/payments/payment-intent
POST /api/payments/retrieve
```

## Prerequisites

- Go 1.26+
- Docker and Docker Compose
- Stripe account (for API keys)

## Quick Start

1. Clone the repository:
```bash
git clone https://github.com/wgod58/go-gin-project.git
cd go-gin-project
```

2. Create a `.env` file:
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
```

3. Start infrastructure services (MySQL + Redis):
```bash
make up
```

4. Run the application:
```bash
go run main.go
```

5. Open Swagger UI:
```
http://localhost:8080/swagger/index.html
```

## Commands

```bash
make up             # start Docker services (MySQL + Redis)
make down           # stop Docker services
make test           # go test ./... -v
make init-db        # create goProject database
make mysql-cli      # connect to MySQL shell
make redis-cli      # connect to Redis shell
make proto          # regenerate protobuf from api/proto/*.proto
make proto-install  # install protoc-gen-go and protoc-gen-go-grpc
make grpc-client    # run gRPC client example
```

## Testing

```bash
go test ./test/... -v
go test ./test/... -v -run TestUserService
```

Tests use `go-sqlmock` for DB interactions and interface-based mocks for Redis and Stripe. See `test/mocks/`.
