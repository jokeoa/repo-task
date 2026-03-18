package memory_test

import (
	"testing"

	"tracker-task/internal/domain"
	"tracker-task/internal/infra/memory"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShipmentRepository_SaveAndRetrieve(t *testing.T) {
	repo := memory.NewShipmentRepository()

	shipment, err := domain.NewShipment("REF-001", "NYC", "LAX")
	require.NoError(t, err)
	shipment.AddUnit(domain.Unit{ID: uuid.New(), Description: "Box A"})

	err = repo.SaveShipment(shipment)
	require.NoError(t, err)

	retrieved, err := repo.GetShipmentByID(shipment.GetID())
	require.NoError(t, err)
	assert.Equal(t, shipment.GetID(), retrieved.GetID())
	assert.Equal(t, shipment.GetReferenceNumber(), retrieved.GetReferenceNumber())
	assert.Equal(t, shipment.GetOrigin(), retrieved.GetOrigin())
	assert.Equal(t, shipment.GetDestination(), retrieved.GetDestination())
	assert.Len(t, retrieved.GetUnits(), 1)
}

func TestShipmentRepository_NotFound(t *testing.T) {
	repo := memory.NewShipmentRepository()

	_, err := repo.GetShipmentByID(uuid.New())

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrShipmentNotFound)
}

func TestShipmentRepository_Overwrite(t *testing.T) {
	repo := memory.NewShipmentRepository()

	shipment, err := domain.NewShipment("REF-001", "NYC", "LAX")
	require.NoError(t, err)

	err = repo.SaveShipment(shipment)
	require.NoError(t, err)

	shipment.AddUnit(domain.Unit{ID: uuid.New(), Description: "Box A"})
	_, err = shipment.AddEvent(domain.StatusPending)
	require.NoError(t, err)

	err = repo.SaveShipment(shipment)
	require.NoError(t, err)

	retrieved, err := repo.GetShipmentByID(shipment.GetID())
	require.NoError(t, err)
	assert.Equal(t, domain.StatusPending, retrieved.GetCurrentStatus())
	assert.Len(t, retrieved.GetUnits(), 1)
}

func TestShipmentRepository_Isolation(t *testing.T) {
	repo := memory.NewShipmentRepository()

	shipment, err := domain.NewShipment("REF-001", "NYC", "LAX")
	require.NoError(t, err)

	err = repo.SaveShipment(shipment)
	require.NoError(t, err)

	// Mutating the original should not affect the stored copy
	shipment.AddUnit(domain.Unit{ID: uuid.New(), Description: "Box A"})

	retrieved, err := repo.GetShipmentByID(shipment.GetID())
	require.NoError(t, err)
	assert.Empty(t, retrieved.GetUnits())
}
