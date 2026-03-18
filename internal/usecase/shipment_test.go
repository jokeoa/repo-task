package usecase_test

import (
	"errors"
	"fmt"
	"testing"

	"tracker-task/internal/domain"
	"tracker-task/internal/usecase"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockShipmentRepository implements domain.ShipmentRepository for testing.
type MockShipmentRepository struct {
	SaveShipmentFn    func(shipment *domain.Shipment) error
	SaveShipmentCalls []*domain.Shipment
	SaveStatuses      []domain.ShipmentStatus
}

func (m *MockShipmentRepository) SaveShipment(shipment *domain.Shipment) error {
	m.SaveShipmentCalls = append(m.SaveShipmentCalls, shipment)
	m.SaveStatuses = append(m.SaveStatuses, shipment.GetCurrentStatus())
	if m.SaveShipmentFn != nil {
		return m.SaveShipmentFn(shipment)
	}
	return nil
}

func (m *MockShipmentRepository) GetShipmentByID(_ uuid.UUID) (domain.Shipment, error) {
	return domain.Shipment{}, nil
}

// MockEventRepository implements domain.EventRepository for testing.
type MockEventRepository struct {
	SaveEventFn    func(event domain.ShipmentEvent) error
	SaveEventCalls []domain.ShipmentEvent
}

func (m *MockEventRepository) SaveEvent(event domain.ShipmentEvent) error {
	m.SaveEventCalls = append(m.SaveEventCalls, event)
	if m.SaveEventFn != nil {
		return m.SaveEventFn(event)
	}
	return nil
}

func (m *MockEventRepository) GetEventsByShipmentID(_ uuid.UUID) ([]domain.ShipmentEvent, error) {
	return nil, nil
}

func newTestService(shipmentRepo *MockShipmentRepository, eventRepo *MockEventRepository) *usecase.ShipmentService {
	return usecase.NewShipmentService(shipmentRepo, eventRepo)
}

func TestNewShipmentService(t *testing.T) {
	shipmentRepo := &MockShipmentRepository{}
	eventRepo := &MockEventRepository{}

	svc := usecase.NewShipmentService(shipmentRepo, eventRepo)
	assert.NotNil(t, svc)
}

func TestCreateShipment(t *testing.T) {
	validUnits := []domain.Unit{
		{ID: uuid.New(), Description: "Box A"},
	}

	tests := []struct {
		name             string
		ref              string
		origin           string
		dest             string
		units            []domain.Unit
		saveShipmentFn   func(*domain.Shipment) error
		saveEventFn      func(domain.ShipmentEvent) error
		expectError      bool
		errContains      string
		errIs            error
		expectSaveCalls  int
		expectEventCalls int
		expectStatus     domain.ShipmentStatus
		expectSavedAs    domain.ShipmentStatus
	}{
		{
			name:        "EmptyUnits",
			ref:         "REF-001",
			origin:      "NYC",
			dest:        "LAX",
			units:       []domain.Unit{},
			expectError: true,
			errContains: "at least one unit",
		},
		{
			name:        "NilUnits",
			ref:         "REF-001",
			origin:      "NYC",
			dest:        "LAX",
			units:       nil,
			expectError: true,
			errContains: "at least one unit",
		},
		{
			name:            "EmptyRef",
			ref:             "",
			origin:          "NYC",
			dest:            "LAX",
			units:           validUnits,
			expectError:     true,
			errIs:           domain.ErrInvalidShipmentData,
			expectSaveCalls: 0,
		},
		{
			name:            "EmptyOrigin",
			ref:             "REF-001",
			origin:          "",
			dest:            "LAX",
			units:           validUnits,
			expectError:     true,
			errIs:           domain.ErrInvalidShipmentData,
			expectSaveCalls: 0,
		},
		{
			name:            "EmptyDest",
			ref:             "REF-001",
			origin:          "NYC",
			dest:            "",
			units:           validUnits,
			expectError:     true,
			errIs:           domain.ErrInvalidShipmentData,
			expectSaveCalls: 0,
		},
		{
			name:   "SaveShipmentError",
			ref:    "REF-001",
			origin: "NYC",
			dest:   "LAX",
			units:  validUnits,
			saveShipmentFn: func(_ *domain.Shipment) error {
				return fmt.Errorf("database connection failed")
			},
			expectError:      true,
			errContains:      "database connection failed",
			expectSaveCalls:  1,
			expectEventCalls: 0,
			expectSavedAs:    domain.StatusPending,
		},
		{
			name:             "ValidInputs",
			ref:              "REF-001",
			origin:           "NYC",
			dest:             "LAX",
			units:            validUnits,
			expectSaveCalls:  1,
			expectEventCalls: 1,
			expectStatus:     domain.StatusPending,
			expectSavedAs:    domain.StatusPending,
		},
		{
			name:   "SaveEventError",
			ref:    "REF-001",
			origin: "NYC",
			dest:   "LAX",
			units:  validUnits,
			saveEventFn: func(_ domain.ShipmentEvent) error {
				return fmt.Errorf("event store unavailable")
			},
			expectError:      true,
			errContains:      "event store unavailable",
			expectSaveCalls:  1,
			expectEventCalls: 1,
			expectSavedAs:    domain.StatusPending,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			shipmentRepo := &MockShipmentRepository{SaveShipmentFn: tc.saveShipmentFn}
			eventRepo := &MockEventRepository{SaveEventFn: tc.saveEventFn}
			svc := newTestService(shipmentRepo, eventRepo)

			shipment, err := svc.CreateShipment(tc.ref, tc.origin, tc.dest, tc.units)

			if tc.expectError {
				require.Error(t, err)
				assert.Nil(t, shipment)

				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				if tc.errIs != nil {
					assert.True(t, errors.Is(err, tc.errIs))
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, shipment)
				assert.Equal(t, tc.expectStatus, shipment.GetCurrentStatus())
				require.Len(t, eventRepo.SaveEventCalls, tc.expectEventCalls)
				assert.Equal(t, domain.StatusPending, eventRepo.SaveEventCalls[0].Status)
				assert.Equal(t, shipment.GetID(), eventRepo.SaveEventCalls[0].ShipmentID)
			}

			assert.Len(t, shipmentRepo.SaveShipmentCalls, tc.expectSaveCalls)
			if tc.expectSaveCalls > 0 {
				assert.Equal(t, tc.expectSavedAs, shipmentRepo.SaveStatuses[0])
			}
			assert.Len(t, eventRepo.SaveEventCalls, tc.expectEventCalls)
		})
	}
}
