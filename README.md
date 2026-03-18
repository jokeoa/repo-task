# Shipment gRPC Microservice

This is a gRPC microservice for handling shipment-related operations. It provides functionalities such as creating shipments, tracking shipments, and managing shipment details.

## Assumptions

- A shipment need something to deliver, so I assume that shipments need more than one unit to be logically complete to do such actions. :3
- A shipment starts in an initial unknown state, and `CreateShipment` applies the first `Pending` event before saving the shipment, so the constructor does not skip the lifecycle transition.
