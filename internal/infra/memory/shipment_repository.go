package memory

import (
	"sync"

	"tracker-task/internal/domain"

	"github.com/google/uuid"
)

type ShipmentRepository struct {
	mu        sync.RWMutex
	shipments map[uuid.UUID]domain.Shipment
}

func NewShipmentRepository() *ShipmentRepository {
	return &ShipmentRepository{
		shipments: make(map[uuid.UUID]domain.Shipment),
	}
}

func (r *ShipmentRepository) SaveShipment(shipment *domain.Shipment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.shipments[shipment.GetID()] = *shipment
	return nil
}

func (r *ShipmentRepository) GetShipmentByID(id uuid.UUID) (domain.Shipment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	shipment, ok := r.shipments[id]
	if !ok {
		return domain.Shipment{}, domain.ErrShipmentNotFound
	}
	return shipment, nil
}
