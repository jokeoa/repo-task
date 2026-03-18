package usecase

import (
	"fmt"

	"tracker-task/internal/domain"

	"github.com/google/uuid"
)

type CreateShipmentInput struct {
	Reference     string
	Origin        string
	Destination   string
	Units         []domain.Unit
	Driver        *uuid.UUID
	Amount        domain.Money
	DriverRevenue domain.Money
}

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

func (s *ShipmentService) CreateShipment(input CreateShipmentInput) (*domain.Shipment, error) {
	if len(input.Units) == 0 {
		return nil, fmt.Errorf("shipment must contain at least one unit")
	}

	shipment, err := domain.NewShipment(input.Reference, input.Origin, input.Destination)
	if err != nil {
		return nil, err
	}

	for _, u := range input.Units {
		shipment.AddUnit(u)
	}

	if input.Driver != nil {
		shipment.SetDriver(*input.Driver)
	}

	if input.Amount != 0 {
		shipment.SetAmount(input.Amount)
	}

	if input.DriverRevenue != 0 {
		shipment.SetDriverRevenue(input.DriverRevenue)
	}

	event, err := shipment.AddEvent(domain.StatusPending)
	if err != nil {
		return nil, err
	}

	if err := s.shipmentRepo.SaveShipment(shipment); err != nil {
		return nil, err
	}

	if err := s.eventRepo.SaveEvent(event); err != nil {
		return nil, err
	}

	return shipment, nil
}

func (s *ShipmentService) AddStatusEvent(shipmentID uuid.UUID, status domain.ShipmentStatus) (*domain.ShipmentEvent, error) {
	shipment, err := s.shipmentRepo.GetShipmentByID(shipmentID)
	if err != nil {
		return nil, err
	}

	event, err := shipment.AddEvent(status)
	if err != nil {
		return nil, err
	}

	if err := s.shipmentRepo.SaveShipment(&shipment); err != nil {
		return nil, err
	}

	if err := s.eventRepo.SaveEvent(event); err != nil {
		return nil, err
	}

	return &event, nil
}

func (s *ShipmentService) GetShipmentByID(id uuid.UUID) (*domain.Shipment, error) {
	shipment, err := s.shipmentRepo.GetShipmentByID(id)
	if err != nil {
		return nil, err
	}
	return &shipment, nil
}

func (s *ShipmentService) GetShipmentHistory(shipmentID uuid.UUID) ([]domain.ShipmentEvent, error) {
	return s.eventRepo.GetEventsByShipmentID(shipmentID)
}
