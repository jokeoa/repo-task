package usecase

import (
	"fmt"
	"tracker-task/internal/domain"
)

type ShipmentService struct {
	shipmentRepo domain.ShipmentRepository
	eventRepo    domain.EventRepository
}

func NewShipmentService(shipmentRepo domain.ShipmentRepository, eventRepo domain.EventRepository) *ShipmentService {
	return &ShipmentService{
		shipmentRepo: shipmentRepo,
		eventRepo:    eventRepo,
	}
}

func (s *ShipmentService) CreateShipment(ref, origin, dest string, units []domain.Unit) (*domain.Shipment, error) {
	if len(units) == 0 {
		return nil, fmt.Errorf("shipment must contain at least one unit")
	}

	shipment, err := domain.NewShipment(ref, origin, dest)

	if err != nil {
		return nil, err
	}

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
