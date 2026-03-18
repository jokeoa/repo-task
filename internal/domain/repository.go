package domain

import "github.com/google/uuid"

type ShipmentRepository interface {
	SaveShipment(shipment Shipment) error
	GetShipmentByID(id uuid.UUID) (Shipment, error)
}

type EventRepository interface {
	SaveEvent(event ShipmentEvent) error
	GetEventsByShipmentID(shipmentID uuid.UUID) ([]ShipmentEvent, error)
}
