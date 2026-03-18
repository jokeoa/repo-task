# Shipment gRPC Microservice

This is a gRPC microservice for handling shipment-related operations. It provides functionalities such as creating shipments, tracking shipments, and managing shipment details.

## Assumptions

- A shipment need something to deliver, so I assume that shipments need more than one unit to be logically complete to do such actions. :3
- A shipment starts in an initial unknown state, and `CreateShipment` applies the first `Pending` event before saving the shipment, so the constructor does not skip the lifecycle transition.
- Driver assignment is optional; a shipment can exist without a driver.
- Monetary values (amount, driver revenue) are stored in cents as `int64` to avoid floating-point precision issues.
- Status strings follow the format: `"Pending"`, `"Picked Up"`, `"In Transit"`, `"Delivered"`, `"Cancelled"` (case-insensitive parsing supported).
