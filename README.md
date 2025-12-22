# Fees API

A REST API for managing bills and line items with workflow automation, built with [Encore](https://encore.dev) and [Temporal](https://temporal.io).

## Features

- **Bill Management**: Create, retrieve, list, and close bills
- **Line Items**: Add detailed line items to bills with descriptions and amounts
- **Currency Support**: Supports USD and Georgian Lari (GEL)
- **Workflow Automation**: Uses Temporal workflows for bill processing
- **PostgreSQL Database**: Persistent storage with migrations
- **RESTful API**: Clean REST endpoints with proper error handling

## Tech Stack

- **Language**: Go 1.25.4
- **Framework**: Encore.dev
- **Workflow Engine**: Temporal SDK
- **Database**: PostgreSQL with pgx driver
- **Testing**: Built-in Go testing with testify

## Prerequisites

- Go 1.25.4 or later
- PostgreSQL database
- Temporal server (for workflow functionality)
- Encore CLI (for development)

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/digitebs/fees-api.git
   cd fees-api
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up your PostgreSQL database and update connection settings in Encore configuration.

4. Start the Temporal server (if running workflows):
   ```bash
   # Using Docker
   docker run -p 7233:7233 temporalio/auto-setup:latest
   ```

5. Run the application:
   ```bash
   encore run
   ```

## API Endpoints

### Bills

- **POST /bills** - Create a new bill
  ```json
  {
    "currency": "USD"
  }
  ```

- **GET /bills** - List all bills (optional status filter)
  - Query parameters: `?status=OPEN` or `?status=CLOSED`

- **GET /bills/:id** - Get a specific bill with line items

- **POST /bills/:id/close** - Close a bill

### Line Items

- **POST /bills/:id/items** - Add a line item to a bill
  ```json
  {
    "amount": 1000,
    "description": "Service fee"
  }
  ```

## Usage Examples

### Creating a Bill

```bash
curl -X POST http://localhost:4000/bills \
  -H "Content-Type: application/json" \
  -d '{"currency": "USD"}'
```

### Adding a Line Item

```bash
curl -X POST http://localhost:4000/bills/{bill-id}/items \
  -H "Content-Type: application/json" \
  -d '{"amount": 5000, "description": "Consulting services"}'
```

### Getting Bill Details

```bash
curl http://localhost:4000/bills/{bill-id}
```

## Data Models

### Bill
```go
type Bill struct {
    ID        string      `json:"id"`
    Status    Status      `json:"status"`    // OPEN or CLOSED
    Total     money.Money `json:"total"`
    CreatedAt time.Time   `json:"created_at"`
    ClosedAt  *time.Time  `json:"closed_at,omitempty"`
}
```

### Line Item
```go
type LineItem struct {
    ID          string      `json:"id"`
    BillID      string      `json:"bill_id"`
    Amount      money.Money `json:"amount"`
    Description string      `json:"description"`
    CreatedAt   time.Time   `json:"created_at"`
}
```

### Money
```go
type Money struct {
    Amount   int64    // Amount in smallest currency unit (cents/tetri)
    Currency Currency // USD or GEL
}
```

## Architecture

The application is structured as follows:

- **bill/**: Main business logic package
  - `api.go`: REST API endpoints
  - `model.go`: Data structures
  - `service.go`: Business logic
  - `repository.go`: Database operations
  - `workflow.go`: Temporal workflow definitions
  - `activities.go`: Temporal activity implementations
  - `worker.go`: Temporal worker setup
  - `db/migrations/`: Database schema migrations

- **money/**: Money handling utilities
  - `money.go`: Currency and money operations

## Database Schema

The application uses PostgreSQL with the following main tables:

- `bills`: Bill records
- `line_items`: Individual bill items

Migrations are located in `bill/db/migrations/`.

## Running Tests

```bash
go test ./...
```

## Development

1. Make sure you have Encore CLI installed
2. Use `encore run` for development with hot reload
3. Use `encore test` to run tests with Encore's testing framework

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

This project is licensed under the MIT License.
