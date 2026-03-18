package domain

import (
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

func (s Shipment) GetDriver() *uuid.UUID {
	return s.driver
}

func (s Shipment) GetAmount() Money {
	return s.amount
}

func (s Shipment) GetDriverRevenue() Money {
	return s.driverRevenue
}

func (s *Shipment) SetDriver(d uuid.UUID) {
	s.driver = &d
}

func (s *Shipment) SetAmount(a Money) {
	s.amount = a
}

func (s *Shipment) SetDriverRevenue(r Money) {
	s.driverRevenue = r
}

func NewShipment(ref, origin, dest string) (*Shipment, error) {
	if ref == "" || origin == "" || dest == "" {
		return nil, ErrInvalidShipmentData
	}
	return &Shipment{
		id:              uuid.New(),
		referenceNumber: ref,
		origin:          origin,
		destination:     dest,
		currentStatus:   StatusUnknown,
		units:           make([]Unit, 0),
	}, nil
}
