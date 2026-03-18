package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Unit struct {
	ID          uuid.UUID
	Description string
}

type Money int64

type Shipment struct {
	id              uuid.UUID
	referenceNumber string
	origin          string
	destination     string
	currentStatus   ShipmentStatus
	driver          *uuid.UUID
	units           []Unit
	amount          Money
	driverRevenue   Money
}

func (s *Shipment) AddUnit(u Unit) {
	s.units = append(s.units, u)
}

func (s Shipment) GetUnits() []Unit {
	return s.units
}
func (s Shipment) GetID() uuid.UUID {
	return s.id
}

func (s Shipment) GetReferenceNumber() string {
	return s.referenceNumber
}
func (s Shipment) GetOrigin() string {
	return s.origin
}
func (s Shipment) GetDestination() string {
	return s.destination
}

func (s Shipment) GetCurrentStatus() ShipmentStatus {
	return s.currentStatus
}

func NewShipment(ref, origin, dest string) *Shipment {
	return &Shipment{
		id:              uuid.New(),
		referenceNumber: ref,
		origin:          origin,
		destination:     dest,
		currentStatus:   StatusPending,
		units:           make([]Unit, 0),
	}
}

const (
	StatusUnknown ShipmentStatus = iota
	StatusPending
	StatusPickedUp
	StatusInTransit
	StatusDelivered
	StatusCancelled
)

type ShipmentStatus int

func (ss ShipmentStatus) GetStatus() (string, error) {
	switch ss {
	case StatusPending:
		return "Pending", nil
	case StatusPickedUp:
		return "Picked Up", nil
	case StatusInTransit:
		return "In Transit", nil
	case StatusDelivered:
		return "Delivered", nil
	case StatusCancelled:
		return "Cancelled", nil
	default:
		return "", fmt.Errorf("invalid shipment status: %d", ss)
	}
}

func (ss ShipmentStatus) CanTransitionTo(next ShipmentStatus) bool {
	switch ss {
	case StatusPending:
		return next == StatusPickedUp || next == StatusCancelled
	case StatusPickedUp:
		return next == StatusInTransit || next == StatusCancelled
	case StatusInTransit:
		return next == StatusDelivered || next == StatusCancelled
	case StatusDelivered:
		return false // a terminal state
	case StatusCancelled:
		return false // a terminal state
	default:
		return false
	}
}

type ShipmentEvent struct {
	ShipmentID uuid.UUID
	Status     ShipmentStatus
	CreatedAt  time.Time
}

func (s *Shipment) AddEvent(ss ShipmentStatus) (ShipmentEvent, error) {
	if !s.currentStatus.CanTransitionTo(ss) {
		return ShipmentEvent{}, fmt.Errorf("invalid status transition from %d to %d", s.currentStatus, ss)
	}

	s.currentStatus = ss

	return ShipmentEvent{
		ShipmentID: s.id,
		Status:     ss,
		CreatedAt:  time.Now(),
	}, nil
}
