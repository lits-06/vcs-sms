# Server Management System

A comprehensive server management system built with Go + Gin framework following Clean Architecture principles. This system manages the on/off status of up to 10,000 servers with monitoring, reporting, and import/export capabilities.

## Features

### Core Functionality
- **Server Management**: Create, view, update, and delete servers
- **Status Monitoring**: Periodic health checks and status updates
- **Filtering & Pagination**: Advanced server listing with filters and sorting
- **Import/Export**: Excel file support for bulk operations
- **Reporting**: Daily email reports with server statistics and uptime metrics

### Technical Features
- **JWT Authentication**: Scope-based authorization for each API endpoint
- **Redis Caching**: Performance optimization with Redis
- **Elasticsearch Integration**: Uptime calculation and analytics
- **OpenAPI Documentation**: Complete API documentation with Swagger
- **PostgreSQL Database**: Secure storage with SQL injection protection
- **Log Management**: File logging with log rotation
- **Unit Testing**: Comprehensive test coverage (>= 90%)

## Architecture

This project follows Clean Architecture principles with the following layers:

```
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── domain/          # Business entities and interfaces
│   ├── usecase/         # Business logic
│   ├── repository/      # Data access layer
│   ├── infrastructure/  # External services (DB, Redis, ES)
│   ├── delivery/http/   # HTTP handlers and routing
│   └── services/        # Application services
└── pkg/                 # Shared packages
```

## Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Redis 7+
- Elasticsearch 8.11+
- Docker & Docker Compose (optional)

## Quick Start

### Using Docker Compose (Recommended)

1. Clone the repository:
```bash
git clone <repository-url>
cd VCS-Checkpoint1
```

2. Copy and configure the environment:
```bash
cp config.example.yaml config.yaml
# Edit config.yaml with your settings
```

3. Start all services:
```bash
make docker-up
```

4. Access the API:
- API: http://localhost:8080
- Swagger Documentation: http://localhost:8080/swagger/index.html

### Manual Setup

1. Install dependencies:
```bash
make deps
```

2. Set up PostgreSQL database:
```sql
CREATE DATABASE server_management;
CREATE USER sms_user WITH PASSWORD 'sms_password';
GRANT ALL PRIVILEGES ON DATABASE server_management TO sms_user;
```

3. Start Redis and Elasticsearch services

4. Configure the application:
```bash
cp config.example.yaml config.yaml
# Edit config.yaml with your database and service settings
```

5. Run the application:
```bash
make run
```

## API Endpoints

### Authentication
All APIs require JWT authentication with appropriate scopes.

### Server Management
- `POST /api/v1/servers` - Create server (scope: `server:create`)
- `GET /api/v1/servers` - List servers with pagination/filtering (scope: `server:read`)
- `GET /api/v1/servers/{id}` - Get server details (scope: `server:read`)
- `PUT /api/v1/servers/{id}` - Update server (scope: `server:update`)
- `DELETE /api/v1/servers/{id}` - Delete server (scope: `server:delete`)

### Import/Export
- `POST /api/v1/servers/import` - Import from Excel (scope: `server:import`)
- `GET /api/v1/servers/export` - Export to Excel (scope: `server:export`)

## Configuration

Key configuration options in `config.yaml`:

```yaml
server:
  port: 8080

database:
  host: localhost
  port: 5432
  name: server_management
  user: sms_user
  password: sms_password

redis:
  host: localhost
  port: 6379

elasticsearch:
  url: http://localhost:9200

jwt:
  secret: your-secret-key
  expiry: 24h

smtp:
  host: smtp.gmail.com
  port: 587
  username: your-email@gmail.com
  password: your-app-password
  admin_email: admin@example.com

monitoring_interval: 60  # seconds
```

## Testing

Run tests with coverage:
```bash
make test-cover
```

The project maintains >= 90% test coverage across all packages.

## Development

### Available Make Commands

- `make build` - Build the application
- `make test` - Run tests
- `make test-cover` - Run tests with coverage report
- `make run` - Run the application
- `make docker-up` - Start all services with Docker
- `make docker-down` - Stop all services
- `make swagger` - Generate Swagger documentation
- `make lint` - Run code linter

### Code Quality

The project uses:
- `golangci-lint` for code linting
- `swag` for API documentation generation
- Comprehensive unit tests
- SQL injection protection with prepared statements
- Input validation and sanitization

## Monitoring & Reporting

### Health Monitoring
- Automatic server health checks every 60 seconds (configurable)
- Status updates stored in PostgreSQL
- Uptime metrics tracked in Elasticsearch

### Daily Reports
- Automated daily email reports sent to administrators
- Includes server count, online/offline status, and average uptime
- HTML formatted email templates

### Logging
- Structured logging with Zap
- Configurable log levels
- Log rotation support
- Request/response logging middleware

## Import/Export Format

### Excel Import Format
The Excel file should have the following columns:
1. Name (string) - Server name
2. Host (string) - Server hostname/IP
3. Port (integer) - Server port
4. Description (string) - Optional description

### Export Format
Exported Excel files include:
- ID, Name, Host, Port, Status, Description
- Creation and update timestamps
- Last check timestamp

## Security

- JWT-based authentication with scope-based authorization
- SQL injection protection using prepared statements
- Input validation and sanitization
- CORS middleware
- Rate limiting (can be added)
- Secure password hashing (for user management)

## Performance Optimization

- Redis caching for frequently accessed data
- Database connection pooling
- Concurrent server health checks using worker pools
- Efficient pagination with database indexes
- Elasticsearch for fast uptime calculations

## Deployment

### Production Deployment

1. Build the Docker image:
```bash
make docker-build
```

2. Deploy using Docker Compose:
```bash
docker-compose -f docker-compose.prod.yml up -d
```

3. Configure environment variables for production
4. Set up SSL/TLS certificates
5. Configure reverse proxy (nginx/traefik)
6. Set up monitoring and alerting

## Contributing

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Ensure all tests pass and coverage >= 90%
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For questions and support, please contact the development team or create an issue in the repository.

- **Server Management**: Create, view, update, delete servers
- **Status Monitoring**: Periodic server status checking and updates  
- **Advanced Filtering**: Filter, pagination, and sorting for server views
- **Import/Export**: Excel file import/export functionality
- **Reporting**: Daily email reports with server statistics and uptime analysis
- **Logging**: Structured logging with log rotation
- **Authentication**: JWT-based authentication with scope-based authorization
- **Caching**: Redis caching for performance optimization
- **Search & Analytics**: Elasticsearch for uptime calculations
- **API Documentation**: OpenAPI/Swagger documentation
- **Testing**: Unit tests with 90%+ coverage
- **Security**: PostgreSQL with SQL injection protection

## Architecture

The project follows Clean Architecture principles:

```
cmd/
├── server/           # Application entry point
internal/
├── domain/           # Business entities and interfaces
├── usecase/          # Business logic layer
├── repository/       # Data access layer
├── delivery/         # HTTP handlers and middleware
├── infrastructure/   # External services (DB, Redis, ES, etc.)
└── config/           # Configuration management
pkg/                  # Shared utilities
docs/                 # API documentation
migrations/           # Database migrations
tests/                # Test files
scripts/              # Deployment and utility scripts
```

## Getting Started

### Prerequisites

- Go 1.24+
- PostgreSQL 13+
- Redis 6+
- Elasticsearch 7+

### Installation

1. Clone the repository
2. Install dependencies: `go mod tidy`
3. Set up environment variables (see `.env.example`)
4. Run migrations: `make migrate-up`
5. Start the server: `make run`

### API Documentation

Access the Swagger UI at: `http://localhost:8080/swagger/index.html`

## Testing

Run tests with coverage:
```bash
make test
make test-coverage
```

## Deployment

Use Docker Compose for local development:
```bash
docker-compose up -d
```
