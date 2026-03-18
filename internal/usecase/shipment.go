package usecase

import (
	"tracker-task/internal/domain"
)

type ShipmentService struct {
	shipmentRepo domain.ShipmentRepository
	eventRepo    domain.EventRepository
}

func (s *ShipmentService) CreateShipment(ref, origin, dest string, units []domain.Unit) (*domain.Shipment, error) {
	if ref == "" || origin == "" || dest == "" || len(units) == 0 {
		return nil, domain.ErrInvalidTransition
	}

	shipment := domain.NewShipment(ref, origin, dest)

	for _, u := range units {
		shipment.AddUnit(u)
	}

	if err := s.shipmentRepo.SaveShipment(shipment); err != nil {
		return nil, err
	}

	event, err := shipment.AddEvent(domain.StatusPending)

	if err != nil {
		return nil, err
	}

	if err := s.eventRepo.SaveEvent(event); err != nil {
		return nil, err
	}

	return shipment, nil
}
