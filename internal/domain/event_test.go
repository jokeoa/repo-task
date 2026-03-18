package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddEvent(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(t *testing.T) *Shipment
		targetStatus   ShipmentStatus
		expectError    bool
		expectedStatus ShipmentStatus
	}{
		{
			name: "Unknown_to_Pending",
			setup: func(t *testing.T) *Shipment {
				s, err := NewShipment("REF-001", "NYC", "LAX")
				require.NoError(t, err)
				return s
			},
			targetStatus:   StatusPending,
			expectedStatus: StatusPending,
		},
		{
			name: "PickedUp_to_InTransit",
			setup: func(t *testing.T) *Shipment {
				s, err := NewShipment("REF-001", "NYC", "LAX")
				require.NoError(t, err)
				_, err = s.AddEvent(StatusPending)
				require.NoError(t, err)
				_, err = s.AddEvent(StatusPickedUp)
				require.NoError(t, err)
				return s
			},
			targetStatus:   StatusInTransit,
			expectedStatus: StatusInTransit,
		},
		{
			name: "InTransit_to_Delivered",
			setup: func(t *testing.T) *Shipment {
				s, err := NewShipment("REF-001", "NYC", "LAX")
				require.NoError(t, err)
				_, err = s.AddEvent(StatusPending)
				require.NoError(t, err)
				_, err = s.AddEvent(StatusPickedUp)
				require.NoError(t, err)
				_, err = s.AddEvent(StatusInTransit)
				require.NoError(t, err)
				return s
			},
			targetStatus:   StatusDelivered,
			expectedStatus: StatusDelivered,
		},
		{
			name: "Pending_to_Cancelled",
			setup: func(t *testing.T) *Shipment {
				s, err := NewShipment("REF-001", "NYC", "LAX")
				require.NoError(t, err)
				_, err = s.AddEvent(StatusPending)
				require.NoError(t, err)
				return s
			},
			targetStatus:   StatusCancelled,
			expectedStatus: StatusCancelled,
		},
		{
			name: "Pending_to_Pending_invalid",
			setup: func(t *testing.T) *Shipment {
				s, err := NewShipment("REF-001", "NYC", "LAX")
				require.NoError(t, err)
				_, err = s.AddEvent(StatusPending)
				require.NoError(t, err)
				return s
			},
			targetStatus:   StatusPending,
			expectError:    true,
			expectedStatus: StatusPending, // unchanged
		},
		{
			name: "Pending_to_InTransit_invalid",
			setup: func(t *testing.T) *Shipment {
				s, err := NewShipment("REF-001", "NYC", "LAX")
				require.NoError(t, err)
				_, err = s.AddEvent(StatusPending)
				require.NoError(t, err)
				return s
			},
			targetStatus:   StatusInTransit,
			expectError:    true,
			expectedStatus: StatusPending,
		},
		{
			name: "Delivered_to_Cancelled_invalid",
			setup: func(t *testing.T) *Shipment {
				s, err := NewShipment("REF-001", "NYC", "LAX")
				require.NoError(t, err)
				_, err = s.AddEvent(StatusPending)
				require.NoError(t, err)
				_, err = s.AddEvent(StatusPickedUp)
				require.NoError(t, err)
				_, err = s.AddEvent(StatusInTransit)
				require.NoError(t, err)
				_, err = s.AddEvent(StatusDelivered)
				require.NoError(t, err)
				return s
			},
			targetStatus:   StatusCancelled,
			expectError:    true,
			expectedStatus: StatusDelivered,
		},
		{
			name: "Cancelled_to_Pending_invalid",
			setup: func(t *testing.T) *Shipment {
				s, err := NewShipment("REF-001", "NYC", "LAX")
				require.NoError(t, err)
				_, err = s.AddEvent(StatusPending)
				require.NoError(t, err)
				_, err = s.AddEvent(StatusCancelled)
				require.NoError(t, err)
				return s
			},
			targetStatus:   StatusPending,
			expectError:    true,
			expectedStatus: StatusCancelled,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			shipment := tc.setup(t)
			before := time.Now()

			event, err := shipment.AddEvent(tc.targetStatus)

			if tc.expectError {
				require.Error(t, err)
				assert.True(t, errors.Is(err, ErrInvalidTransition))
				assert.Equal(t, ShipmentEvent{}, event)
				assert.Equal(t, tc.expectedStatus, shipment.GetCurrentStatus())
				return
			}

			require.NoError(t, err)
			assert.Equal(t, shipment.GetID(), event.ShipmentID)
			assert.Equal(t, tc.targetStatus, event.Status)
			assert.WithinDuration(t, time.Now(), event.CreatedAt, time.Second)
			assert.False(t, event.CreatedAt.Before(before))
			assert.Equal(t, tc.expectedStatus, shipment.GetCurrentStatus())
		})
	}
}

func TestAddEvent_FullLifecycle(t *testing.T) {
	shipment, err := NewShipment("REF-LIFE", "NYC", "LAX")
	require.NoError(t, err)
	assert.Equal(t, StatusUnknown, shipment.GetCurrentStatus())

	steps := []ShipmentStatus{StatusPending, StatusPickedUp, StatusInTransit, StatusDelivered}

	for _, status := range steps {
		event, err := shipment.AddEvent(status)
		require.NoError(t, err)
		assert.Equal(t, shipment.GetID(), event.ShipmentID)
		assert.Equal(t, status, event.Status)
		assert.Equal(t, status, shipment.GetCurrentStatus())
		assert.NotEqual(t, uuid.Nil, event.ShipmentID)
	}
}
