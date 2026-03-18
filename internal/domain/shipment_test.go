package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewShipment(t *testing.T) {
	tests := []struct {
		name        string
		ref         string
		origin      string
		dest        string
		expectError error
	}{
		{name: "ValidInputs", ref: "REF-001", origin: "NYC", dest: "LAX"},
		{name: "EmptyRef", ref: "", origin: "NYC", dest: "LAX", expectError: ErrInvalidShipmentData},
		{name: "EmptyOrigin", ref: "REF-001", origin: "", dest: "LAX", expectError: ErrInvalidShipmentData},
		{name: "EmptyDest", ref: "REF-001", origin: "NYC", dest: "", expectError: ErrInvalidShipmentData},
		{name: "AllEmpty", ref: "", origin: "", dest: "", expectError: ErrInvalidShipmentData},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			shipment, err := NewShipment(tc.ref, tc.origin, tc.dest)

			if tc.expectError != nil {
				require.ErrorIs(t, err, tc.expectError)
				assert.Nil(t, shipment)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, shipment)

			assert.Equal(t, tc.ref, shipment.GetReferenceNumber())
			assert.Equal(t, tc.origin, shipment.GetOrigin())
			assert.Equal(t, tc.dest, shipment.GetDestination())
			assert.Equal(t, StatusUnknown, shipment.GetCurrentStatus())
			assert.NotEqual(t, uuid.Nil, shipment.GetID())
			assert.NotNil(t, shipment.GetUnits())
			assert.Empty(t, shipment.GetUnits())
		})
	}
}

func TestAddUnit(t *testing.T) {
	t.Run("SingleUnit", func(t *testing.T) {
		shipment, err := NewShipment("REF-001", "NYC", "LAX")
		require.NoError(t, err)

		unit := Unit{ID: uuid.New(), Description: "Box A"}
		shipment.AddUnit(unit)

		units := shipment.GetUnits()
		require.Len(t, units, 1)
		assert.Equal(t, unit.ID, units[0].ID)
		assert.Equal(t, unit.Description, units[0].Description)
	})

	t.Run("MultipleUnits", func(t *testing.T) {
		shipment, err := NewShipment("REF-002", "NYC", "LAX")
		require.NoError(t, err)

		unit1 := Unit{ID: uuid.New(), Description: "Box A"}
		unit2 := Unit{ID: uuid.New(), Description: "Box B"}
		unit3 := Unit{ID: uuid.New(), Description: "Box C"}

		shipment.AddUnit(unit1)
		shipment.AddUnit(unit2)
		shipment.AddUnit(unit3)

		units := shipment.GetUnits()
		require.Len(t, units, 3)
		assert.Equal(t, unit1.Description, units[0].Description)
		assert.Equal(t, unit2.Description, units[1].Description)
		assert.Equal(t, unit3.Description, units[2].Description)
	})
}

func TestNewShipment_UniqueIDs(t *testing.T) {
	s1, err := NewShipment("REF-001", "NYC", "LAX")
	require.NoError(t, err)

	s2, err := NewShipment("REF-002", "NYC", "LAX")
	require.NoError(t, err)

	assert.NotEqual(t, s1.GetID(), s2.GetID())
}
