package domain

type ShipmentRepository interface {
	SaveShipment(shipment Shipment) error
	GetShipmentByID(id string) (Shipment, error)
}
