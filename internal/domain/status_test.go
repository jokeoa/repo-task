package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetStatus(t *testing.T) {
	tests := []struct {
		name        string
		status      ShipmentStatus
		expected    string
		expectError bool
		errContains string
	}{
		{name: "Unknown", status: StatusUnknown, expected: "Unknown"},
		{name: "Pending", status: StatusPending, expected: "Pending"},
		{name: "PickedUp", status: StatusPickedUp, expected: "Picked Up"},
		{name: "InTransit", status: StatusInTransit, expected: "In Transit"},
		{name: "Delivered", status: StatusDelivered, expected: "Delivered"},
		{name: "Cancelled", status: StatusCancelled, expected: "Cancelled"},
		{name: "InvalidPositive", status: ShipmentStatus(99), expectError: true, errContains: "invalid shipment status"},
		{name: "InvalidNegative", status: ShipmentStatus(-1), expectError: true, errContains: "invalid shipment status"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.status.GetStatus()

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
				assert.Empty(t, result)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		from     ShipmentStatus
		to       ShipmentStatus
		expected bool
	}{
		// Valid transitions from Pending
		{name: "Pending_to_PickedUp", from: StatusPending, to: StatusPickedUp, expected: true},
		{name: "Pending_to_Cancelled", from: StatusPending, to: StatusCancelled, expected: true},

		// Valid transitions from PickedUp
		{name: "PickedUp_to_InTransit", from: StatusPickedUp, to: StatusInTransit, expected: true},
		{name: "PickedUp_to_Cancelled", from: StatusPickedUp, to: StatusCancelled, expected: true},

		// Valid transitions from InTransit
		{name: "InTransit_to_Delivered", from: StatusInTransit, to: StatusDelivered, expected: true},
		{name: "InTransit_to_Cancelled", from: StatusInTransit, to: StatusCancelled, expected: true},

		// Valid: initial state must enter Pending first
		{name: "Unknown_to_Pending", from: StatusUnknown, to: StatusPending, expected: true},

		// Invalid: Pending self-transition after initialization
		{name: "Pending_to_Pending", from: StatusPending, to: StatusPending, expected: false},

		// Invalid: skip states
		{name: "Pending_to_InTransit", from: StatusPending, to: StatusInTransit, expected: false},
		{name: "Pending_to_Delivered", from: StatusPending, to: StatusDelivered, expected: false},
		{name: "Pending_to_Unknown", from: StatusPending, to: StatusUnknown, expected: false},

		// Invalid: self-transitions
		{name: "PickedUp_to_PickedUp", from: StatusPickedUp, to: StatusPickedUp, expected: false},
		{name: "InTransit_to_InTransit", from: StatusInTransit, to: StatusInTransit, expected: false},
		{name: "Delivered_to_Delivered", from: StatusDelivered, to: StatusDelivered, expected: false},
		{name: "Cancelled_to_Cancelled", from: StatusCancelled, to: StatusCancelled, expected: false},

		// Invalid: backward transitions
		{name: "PickedUp_to_Pending", from: StatusPickedUp, to: StatusPending, expected: false},
		{name: "InTransit_to_Pending", from: StatusInTransit, to: StatusPending, expected: false},
		{name: "InTransit_to_PickedUp", from: StatusInTransit, to: StatusPickedUp, expected: false},

		// Terminal states: Delivered cannot transition
		{name: "Delivered_to_Pending", from: StatusDelivered, to: StatusPending, expected: false},
		{name: "Delivered_to_Cancelled", from: StatusDelivered, to: StatusCancelled, expected: false},

		// Terminal states: Cancelled cannot transition
		{name: "Cancelled_to_Pending", from: StatusCancelled, to: StatusPending, expected: false},
		{name: "Cancelled_to_PickedUp", from: StatusCancelled, to: StatusPickedUp, expected: false},

		// Default branch: Unknown and invalid statuses
		{name: "Unknown_to_Cancelled", from: StatusUnknown, to: StatusCancelled, expected: false},
		{name: "Unknown_to_PickedUp", from: StatusUnknown, to: StatusPickedUp, expected: false},
		{name: "Invalid99_to_Pending", from: ShipmentStatus(99), to: StatusPending, expected: false},
		{name: "Invalid99_to_PickedUp", from: ShipmentStatus(99), to: StatusPickedUp, expected: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.from.CanTransitionTo(tc.to)
			assert.Equal(t, tc.expected, result)
		})
	}
}
