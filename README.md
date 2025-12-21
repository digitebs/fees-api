# Fees API

A backend service for managing bills and fees, built with [Encore](https://encore.dev/) and [Temporal](https://temporal.io/). This API provides functionality to create, manage, and process bills with line items, supporting multiple currencies.

## Features

- **Bill Management**: Create, retrieve, and close bills
- **Line Items**: Add items to bills with amounts and descriptions
- **Multi-Currency Support**: Supports USD and GEL currencies
- **Workflow Processing**: Uses Temporal workflows for bill processing and finalization
- **Database Integration**: Uses PostgreSQL with database migrations
- **REST API**: Public API endpoints for bill operations

## Technology Stack

- **Framework**: Encore v1.52.1
- **Workflow Engine**: Temporal SDK v1.38.0
- **Language**: Go 1.25.4
- **Database**: PostgreSQL
- **Dependencies**:
  - `encore.dev` - Backend framework
  - `go.temporal.io/sdk` - Temporal workflow SDK
  - `github.com/google/uuid` - UUID generation
  - `github.com/jackc/pgx/v5` - PostgreSQL driver

## Project Structure

```
fees-api/
├── bill/                    # Bill service
│   ├── activities.go       # Temporal activities
│   ├── api.go              # Public API endpoints
│   ├── config.cue          # Configuration
│   ├── config.go           # Configuration loader
│   ├── model.go            # Data models
│   ├── repository.go       # Database operations
│   ├── service.go          # Business logic
│   ├── signals.go          # Temporal signals
│   ├── workflow.go         # Temporal workflows
│   ├── worker.go           # Temporal worker setup
│   ├── bill_test.go        # Unit tests
│   ├── signal_test.go      # Signal tests
│   └── db/migrations/      # Database migrations
├── money/                   # Money handling package
│   ├── money.go            # Money types and operations
│   └── money_test.go       # Money tests
├── temporal/                # Temporal client setup
│   └── client.go           # Temporal client configuration
├── encore.app              # Encore app configuration
├── go.mod                  # Go module file
├── go.sum                  # Go dependencies
└── README.md               # This file
```

## API Endpoints

### Bills

- `POST /bills` - Create a new bill
  - Body: `{"currency": "USD"}` or `{"currency": "GEL"}`
- `GET /bills` - List all bills (optional query param: `status=OPEN|CLOSED`)
- `GET /bills/:id` - Get bill details with line items
- `POST /bills/:id/close` - Close a bill
- `POST /bills/:id/items` - Add a line item to a bill
  - Body: `{"amount": 1234, "description": "Service fee"}`

### Health Check

- `GET /health` - Service health check

## Getting Started

### Prerequisites

- Go 1.25.4 or later
- PostgreSQL database
- Temporal server running locally (default: localhost:7233)

### Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd fees-api
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up your database and update configuration in `bill/config.cue` if needed.

4. Start the Temporal server (if not already running):
   ```bash
   # Using Temporal CLI
   temporal server start-dev
   ```

5. Run the Encore application:
   ```bash
   encore run
   ```

The API will be available at `http://localhost:4000`.

## Configuration

Configuration is managed through Encore's config system:

- **Temporal Server**: Configured in `bill/config.cue` (default: localhost:7233)
- Database connection is handled automatically by Encore

## Database Migrations

Database schema is managed through migrations in `bill/db/migrations/`:

- `001_create_bills.up.sql` - Creates bills table
- `002_create_line_items.up.sql` - Creates line items table

Migrations are applied automatically by Encore when the service starts.

## Workflows and Activities

The service uses Temporal workflows for bill processing:

- **BillWorkflow**: Main workflow for bill lifecycle
- **FinalizeBillActivity**: Activity to finalize bill amounts
- **PersistLineItemActivity**: Activity to persist line items to database

## Testing

Run tests with:
```bash
go test ./...
```

## Development

- Use `encore test` to run Encore-specific tests
- Use `encore run` for local development with hot reload
- Access Encore dashboard at `http://localhost:4000` for API documentation and monitoring

## Deployment

Deploy to Encore's cloud platform:
```bash
encore deploy
```

Or build for production:
```bash
encore build
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

This project is licensed under the MIT License.
