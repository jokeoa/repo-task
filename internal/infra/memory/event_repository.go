package memory

import (
	"sync"

	"tracker-task/internal/domain"

	"github.com/google/uuid"
)

type EventRepository struct {
	mu     sync.RWMutex
	events map[uuid.UUID][]domain.ShipmentEvent
}

func NewEventRepository() *EventRepository {
	return &EventRepository{
		events: make(map[uuid.UUID][]domain.ShipmentEvent),
	}
}

func (r *EventRepository) SaveEvent(event domain.ShipmentEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.events[event.ShipmentID] = append(r.events[event.ShipmentID], event)
	return nil
}

func (r *EventRepository) GetEventsByShipmentID(shipmentID uuid.UUID) ([]domain.ShipmentEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	events, ok := r.events[shipmentID]
	if !ok {
		return []domain.ShipmentEvent{}, nil
	}

	result := make([]domain.ShipmentEvent, len(events))
	copy(result, events)
	return result, nil
}
