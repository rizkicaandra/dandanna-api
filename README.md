# Dandanna API

A RESTful API built with Go 1.25+ using standard library and clean architecture principles.

## Project Structure

```
.
├── cmd/
│   └── api/              # Application entry point
├── internal/
│   ├── api/              # HTTP layer (handlers, middleware, router)
│   ├── application/      # Application services layer
│   ├── domain/           # Domain layer (entities, repositories)
│   └── infrastructure/   # External dependencies (config, logger, DB)
├── pkg/                  # Public reusable packages
├── migrations/           # Database migrations
├── scripts/              # Build and deployment scripts
├── Makefile              # Build commands
├── Dockerfile            # Container image
└── go.mod               # Go modules
```

## Getting Started

### Prerequisites

- Go 1.25 or higher
- PostgreSQL 16+
- Redis 8+ (optional)
- Docker (optional)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd dandanna-api
```

2. Copy environment variables:
```bash
cp .env.example .env
```

3. Install dependencies:
```bash
make deps
```

### Running the Application

#### Development mode
```bash
make run
```

#### Build binary
```bash
make build
./bin/dandanna-api
```

#### Docker
```bash
make docker-build
make docker-run
```

### API Endpoints

#### Health Check
```
GET /api/health
```

Response:
```json
{
  "success": true,
  "meta": {
    "requestId": "ee81caa76db4974df41cbb8fb23b416d",
    "timestamp": "2026-04-14T08:30:00.000Z"
  },
  "data": {
    "status": "healthy",
    "version": "1.0.0"
  },
  "errors": []
}
```

---

### Response Format

All endpoints use a consistent JSON envelope.

#### Success
```json
{
  "success": true,
  "meta": {
    "requestId": "ee81caa76db4974df41cbb8fb23b416d",
    "timestamp": "2026-04-14T08:30:00.000Z"
  },
  "data": { },
  "errors": []
}
```

#### Error
```json
{
  "success": false,
  "meta": {
    "requestId": "z9y8x7w6v5u4t3s2r1q0",
    "timestamp": "2026-04-14T08:31:00.000Z"
  },
  "data": null,
  "errors": [
    {
      "field": "bookingDate",
      "code": "DATE_IN_THE_PAST",
      "meta": {},
      "message": "Booking date must be in the future."
    },
    {
      "field": "balance",
      "code": "INSUFFICIENT_BALANCE",
      "meta": {
        "required": 350000,
        "current": 100000,
        "currency": "IDR"
      },
      "message": "Balance is not enough to complete this booking."
    }
  ]
}
```

#### Rules
| Field | Success | Error |
|---|---|---|
| `success` | `true` | `false` |
| `meta.requestId` | X-Request-ID from context | X-Request-ID from context |
| `meta.timestamp` | RFC3339Nano UTC | RFC3339Nano UTC |
| `data` | response payload | `null` |
| `errors` | `[]` (never null) | one or more error items |
| `errors[].field` | — | omitted for non-field errors |
| `errors[].code` | — | SCREAMING\_SNAKE\_CASE |
| `errors[].meta` | — | `{}` when no extra context |

## Development

### Available Make Commands

```bash
make help           # Show all available commands
make build          # Build the application
make run            # Run the application
make test           # Run tests
make test-coverage  # Run tests with coverage
make lint           # Run linters
make fmt            # Format code
make tidy           # Tidy go modules
make clean          # Clean build artifacts
```

### Testing

Run all tests:
```bash
make test
```

Run tests with coverage:
```bash
make test-coverage
```

### Linting

Run linters:
```bash
make lint
```

## Architecture

This project follows **Clean Architecture** principles with clear separation of concerns:

- **Domain Layer**: Core business logic, entities, and interfaces
- **Application Layer**: Use cases and business rules orchestration
- **Infrastructure Layer**: External dependencies (database, cache, logging)
- **API Layer**: HTTP handlers, middleware, and routing

### Key Design Principles

- **Dependency Inversion**: Depend on interfaces, not implementations
- **Repository Pattern**: Abstract data access behind interfaces
- **DTO Pattern**: Separate internal models from API contracts
- **Context Propagation**: All I/O operations accept and propagate context
- **Graceful Shutdown**: Proper handling of SIGTERM/SIGINT

## Configuration

Configuration is managed through environment variables. See `.env.example` for all available options.

Key configuration:
- `SERVER_PORT`: HTTP server port (default: 8080)
- `APP_ENV`: Environment (development, staging, production)
- `LOG_LEVEL`: Logging level (debug, info, warn, error)

## License

[Your License]
