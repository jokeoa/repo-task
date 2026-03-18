package usecase

import "tracker-task/internal/domain"

type ShipmentService struct {
	shipmentRepo domain.ShipmentRepository
	eventRepo    domain.EventRepository
}

func (s *ShipmentService) CreateShipment(ref, origin, dest string, unit []domain.Unit) (*domain.Shipment, error) {
	if ref == "" || origin == "" || dest == "" || len(unit) == 0 {
		return nil, domain.ErrInvalidShipmentData
	}
	shipment := domain.NewShipment(ref, origin, dest)
	return shipment, nil

}
