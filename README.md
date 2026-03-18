# Shipment gRPC Microservice

A gRPC microservice for handling shipment-related operations using clean architecture principles in Go.

## How to Run

```bash
# Run the service
go run main.go

# Run with custom port
PORT=8080 go run main.go

# Run all tests
go test ./...

# Run with Docker
docker compose up --build
```

## Architecture

```
main.go                          # Server wiring & startup
├── internal/
│   ├── domain/                  # Domain layer (entities, value objects, interfaces)
│   │   ├── shipment.go          # Shipment aggregate root
│   │   ├── event.go             # ShipmentEvent value object
│   │   ├── status.go            # ShipmentStatus state machine
│   │   ├── repository.go        # Repository interfaces
│   │   └── errors.go            # Domain error sentinels
│   ├── usecase/                 # Application layer (use cases)
│   │   └── shipment.go          # ShipmentService orchestration
│   ├── infra/memory/            # Infrastructure layer (persistence)
│   │   ├── shipment_repository.go
│   │   └── event_repository.go
│   └── handler/grpc/            # Interface layer (gRPC transport)
│       ├── handler.go           # gRPC endpoint implementations
│       └── mapper.go            # Proto <-> Domain mappers
├── proto/                       # Protocol Buffers definition
│   └── shipment.proto
└── gen/proto/                   # Generated gRPC code
```

## API

| RPC | Description |
|-----|-------------|
| `CreateShipment` | Create a new shipment with units, optional driver, and monetary values |
| `AddStatusEvent` | Advance shipment status (validates state machine transitions) |
| `GetShipmentByID` | Retrieve a shipment by its ID |
| `GetShipmentHistory` | Get the full event history for a shipment |

### Status Lifecycle

```
Unknown → Pending → Picked Up → In Transit → Delivered
                ↘              ↘              ↘
                  Cancelled     Cancelled      Cancelled
```

Terminal states: `Delivered`, `Cancelled` (no further transitions allowed).

## Design Decisions

- **State Machine**: Status transitions are enforced at the domain level. Invalid transitions return `ErrInvalidTransition`.
- **Event Sourcing Foundation**: Every status change creates a `ShipmentEvent` with a timestamp, enabling full audit trails.
- **In-Memory Repositories**: Thread-safe (`sync.RWMutex`) in-memory storage. Stores by value to prevent external mutation.
- **CreateShipmentInput Pattern**: Structured input for `CreateShipment` to cleanly handle optional fields (driver, amounts).
- **Cents for Money**: Monetary values stored as `int64` cents to avoid floating-point precision issues.

## Assumptions

- A shipment needs something to deliver, so shipments require at least one unit.
- A shipment starts in an initial unknown state, and `CreateShipment` applies the first `Pending` event before saving the shipment, so the constructor does not skip the lifecycle transition.
- Driver assignment is optional; a shipment can exist without a driver.
- Monetary values (amount, driver revenue) are stored in cents as `int64` to avoid floating-point precision issues.
- Status strings follow the format: `"Pending"`, `"Picked Up"`, `"In Transit"`, `"Delivered"`, `"Cancelled"` (case-insensitive parsing supported).
