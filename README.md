# Go Payment Service

A robust payment service built with Go, integrating Stripe for payment processing, MySQL for data persistence, and Redis for caching. This service provides RESTful APIs for user management and payment processing.

## Tech Stack

- **Go** (1.21+)
- **Gin** - Web framework
- **GORM** - ORM for MySQL
- **Redis** - Caching layer
- **Stripe** - Payment processing
- **MySQL** - Primary database
- **Docker** - Containerization
- **Swagger** - API documentation
- **JWT** - Authentication

## Features

- User Management (CRUD operations)
- Payment Processing with Stripe
- Redis Caching
- Swagger Documentation
- Docker Containerization
- Comprehensive Test Coverage


## Prerequisites

- Go 1.21+
- Docker and Docker Compose
- Stripe Account (for API keys)

## Quick Start

1.Clone the repository:

```bash
git clone https://github.com/zeddpai/go-gin-project.git
```

2.Create all the service

```bash
docker-compose up -d
```

3.Run the application

```bash
go run main.go
```
