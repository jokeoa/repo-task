package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ShipmentEvent struct {
	ShipmentID uuid.UUID
	Status     ShipmentStatus
	CreatedAt  time.Time
}

func (s *Shipment) AddEvent(ss ShipmentStatus) (ShipmentEvent, error) {
	if !s.currentStatus.CanTransitionTo(ss) {
		return ShipmentEvent{}, fmt.Errorf("%w: from %d to %d", ErrInvalidTransition, s.currentStatus, ss)
	}

	s.currentStatus = ss

	return ShipmentEvent{
		ShipmentID: s.id,
		Status:     ss,
		CreatedAt:  time.Now(),
	}, nil
}
