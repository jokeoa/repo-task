package memory_test

import (
	"testing"
	"time"

	"tracker-task/internal/domain"
	"tracker-task/internal/infra/memory"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventRepository_SaveAndRetrieve(t *testing.T) {
	repo := memory.NewEventRepository()
	shipmentID := uuid.New()

	event := domain.ShipmentEvent{
		ShipmentID: shipmentID,
		Status:     domain.StatusPending,
		CreatedAt:  time.Now(),
	}

	err := repo.SaveEvent(event)
	require.NoError(t, err)

	events, err := repo.GetEventsByShipmentID(shipmentID)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, domain.StatusPending, events[0].Status)
	assert.Equal(t, shipmentID, events[0].ShipmentID)
}

func TestEventRepository_Empty(t *testing.T) {
	repo := memory.NewEventRepository()

	events, err := repo.GetEventsByShipmentID(uuid.New())
	require.NoError(t, err)
	assert.Empty(t, events)
}

func TestEventRepository_MultipleEvents(t *testing.T) {
	repo := memory.NewEventRepository()
	shipmentID := uuid.New()

	statuses := []domain.ShipmentStatus{
		domain.StatusPending,
		domain.StatusPickedUp,
		domain.StatusInTransit,
	}

	for _, status := range statuses {
		err := repo.SaveEvent(domain.ShipmentEvent{
			ShipmentID: shipmentID,
			Status:     status,
			CreatedAt:  time.Now(),
		})
		require.NoError(t, err)
	}

	events, err := repo.GetEventsByShipmentID(shipmentID)
	require.NoError(t, err)
	require.Len(t, events, 3)

	for i, status := range statuses {
		assert.Equal(t, status, events[i].Status)
	}
}

func TestEventRepository_Isolation(t *testing.T) {
	repo := memory.NewEventRepository()
	shipmentID1 := uuid.New()
	shipmentID2 := uuid.New()

	err := repo.SaveEvent(domain.ShipmentEvent{
		ShipmentID: shipmentID1,
		Status:     domain.StatusPending,
		CreatedAt:  time.Now(),
	})
	require.NoError(t, err)

	err = repo.SaveEvent(domain.ShipmentEvent{
		ShipmentID: shipmentID2,
		Status:     domain.StatusPickedUp,
		CreatedAt:  time.Now(),
	})
	require.NoError(t, err)

	events1, err := repo.GetEventsByShipmentID(shipmentID1)
	require.NoError(t, err)
	assert.Len(t, events1, 1)
	assert.Equal(t, domain.StatusPending, events1[0].Status)

	events2, err := repo.GetEventsByShipmentID(shipmentID2)
	require.NoError(t, err)
	assert.Len(t, events2, 1)
	assert.Equal(t, domain.StatusPickedUp, events2[0].Status)
}
