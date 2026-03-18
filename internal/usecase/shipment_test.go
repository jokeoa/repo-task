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
	SaveShipmentFn      func(shipment *domain.Shipment) error
	GetShipmentByIDFn   func(id uuid.UUID) (domain.Shipment, error)
	SaveShipmentCalls   []*domain.Shipment
	SaveStatuses        []domain.ShipmentStatus
}

func (m *MockShipmentRepository) SaveShipment(shipment *domain.Shipment) error {
	m.SaveShipmentCalls = append(m.SaveShipmentCalls, shipment)
	m.SaveStatuses = append(m.SaveStatuses, shipment.GetCurrentStatus())
	if m.SaveShipmentFn != nil {
		return m.SaveShipmentFn(shipment)
	}
	return nil
}

func (m *MockShipmentRepository) GetShipmentByID(id uuid.UUID) (domain.Shipment, error) {
	if m.GetShipmentByIDFn != nil {
		return m.GetShipmentByIDFn(id)
	}
	return domain.Shipment{}, domain.ErrShipmentNotFound
}

// MockEventRepository implements domain.EventRepository for testing.
type MockEventRepository struct {
	SaveEventFn              func(event domain.ShipmentEvent) error
	GetEventsByShipmentIDFn  func(id uuid.UUID) ([]domain.ShipmentEvent, error)
	SaveEventCalls           []domain.ShipmentEvent
}

func (m *MockEventRepository) SaveEvent(event domain.ShipmentEvent) error {
	m.SaveEventCalls = append(m.SaveEventCalls, event)
	if m.SaveEventFn != nil {
		return m.SaveEventFn(event)
	}
	return nil
}

func (m *MockEventRepository) GetEventsByShipmentID(id uuid.UUID) ([]domain.ShipmentEvent, error) {
	if m.GetEventsByShipmentIDFn != nil {
		return m.GetEventsByShipmentIDFn(id)
	}
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
		input            usecase.CreateShipmentInput
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
			name: "EmptyUnits",
			input: usecase.CreateShipmentInput{
				Reference: "REF-001", Origin: "NYC", Destination: "LAX",
				Units: []domain.Unit{},
			},
			expectError: true,
			errContains: "at least one unit",
		},
		{
			name: "NilUnits",
			input: usecase.CreateShipmentInput{
				Reference: "REF-001", Origin: "NYC", Destination: "LAX",
				Units: nil,
			},
			expectError: true,
			errContains: "at least one unit",
		},
		{
			name: "EmptyRef",
			input: usecase.CreateShipmentInput{
				Reference: "", Origin: "NYC", Destination: "LAX",
				Units: validUnits,
			},
			expectError:     true,
			errIs:           domain.ErrInvalidShipmentData,
			expectSaveCalls: 0,
		},
		{
			name: "EmptyOrigin",
			input: usecase.CreateShipmentInput{
				Reference: "REF-001", Origin: "", Destination: "LAX",
				Units: validUnits,
			},
			expectError:     true,
			errIs:           domain.ErrInvalidShipmentData,
			expectSaveCalls: 0,
		},
		{
			name: "EmptyDest",
			input: usecase.CreateShipmentInput{
				Reference: "REF-001", Origin: "NYC", Destination: "",
				Units: validUnits,
			},
			expectError:     true,
			errIs:           domain.ErrInvalidShipmentData,
			expectSaveCalls: 0,
		},
		{
			name: "SaveShipmentError",
			input: usecase.CreateShipmentInput{
				Reference: "REF-001", Origin: "NYC", Destination: "LAX",
				Units: validUnits,
			},
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
			name: "ValidInputs",
			input: usecase.CreateShipmentInput{
				Reference: "REF-001", Origin: "NYC", Destination: "LAX",
				Units: validUnits,
			},
			expectSaveCalls:  1,
			expectEventCalls: 1,
			expectStatus:     domain.StatusPending,
			expectSavedAs:    domain.StatusPending,
		},
		{
			name: "SaveEventError",
			input: usecase.CreateShipmentInput{
				Reference: "REF-001", Origin: "NYC", Destination: "LAX",
				Units: validUnits,
			},
			saveEventFn: func(_ domain.ShipmentEvent) error {
				return fmt.Errorf("event store unavailable")
			},
			expectError:      true,
			errContains:      "event store unavailable",
			expectSaveCalls:  1,
			expectEventCalls: 1,
			expectSavedAs:    domain.StatusPending,
		},
		{
			name: "WithDriverAndAmount",
			input: func() usecase.CreateShipmentInput {
				driverID := uuid.New()
				return usecase.CreateShipmentInput{
					Reference: "REF-002", Origin: "NYC", Destination: "LAX",
					Units:         validUnits,
					Driver:        &driverID,
					Amount:        5000,
					DriverRevenue: 1500,
				}
			}(),
			expectSaveCalls:  1,
			expectEventCalls: 1,
			expectStatus:     domain.StatusPending,
			expectSavedAs:    domain.StatusPending,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			shipmentRepo := &MockShipmentRepository{SaveShipmentFn: tc.saveShipmentFn}
			eventRepo := &MockEventRepository{SaveEventFn: tc.saveEventFn}
			svc := newTestService(shipmentRepo, eventRepo)

			shipment, err := svc.CreateShipment(tc.input)

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

func TestAddStatusEvent(t *testing.T) {
	// Helper to create a shipment in Pending state
	createPendingShipment := func() *domain.Shipment {
		s, _ := domain.NewShipment("REF-001", "NYC", "LAX")
		s.AddUnit(domain.Unit{ID: uuid.New(), Description: "Box"})
		s.AddEvent(domain.StatusPending)
		return s
	}

	t.Run("ValidTransition", func(t *testing.T) {
		pending := createPendingShipment()
		shipmentRepo := &MockShipmentRepository{
			GetShipmentByIDFn: func(_ uuid.UUID) (domain.Shipment, error) {
				return *pending, nil
			},
		}
		eventRepo := &MockEventRepository{}
		svc := newTestService(shipmentRepo, eventRepo)

		event, err := svc.AddStatusEvent(pending.GetID(), domain.StatusPickedUp)

		require.NoError(t, err)
		require.NotNil(t, event)
		assert.Equal(t, domain.StatusPickedUp, event.Status)
		assert.Len(t, shipmentRepo.SaveShipmentCalls, 1)
		assert.Len(t, eventRepo.SaveEventCalls, 1)
	})

	t.Run("InvalidTransition", func(t *testing.T) {
		pending := createPendingShipment()
		shipmentRepo := &MockShipmentRepository{
			GetShipmentByIDFn: func(_ uuid.UUID) (domain.Shipment, error) {
				return *pending, nil
			},
		}
		eventRepo := &MockEventRepository{}
		svc := newTestService(shipmentRepo, eventRepo)

		event, err := svc.AddStatusEvent(pending.GetID(), domain.StatusDelivered)

		require.Error(t, err)
		assert.Nil(t, event)
		assert.True(t, errors.Is(err, domain.ErrInvalidTransition))
		assert.Empty(t, shipmentRepo.SaveShipmentCalls)
		assert.Empty(t, eventRepo.SaveEventCalls)
	})

	t.Run("ShipmentNotFound", func(t *testing.T) {
		shipmentRepo := &MockShipmentRepository{}
		eventRepo := &MockEventRepository{}
		svc := newTestService(shipmentRepo, eventRepo)

		event, err := svc.AddStatusEvent(uuid.New(), domain.StatusPickedUp)

		require.Error(t, err)
		assert.Nil(t, event)
		assert.True(t, errors.Is(err, domain.ErrShipmentNotFound))
	})

	t.Run("SaveShipmentError", func(t *testing.T) {
		pending := createPendingShipment()
		shipmentRepo := &MockShipmentRepository{
			GetShipmentByIDFn: func(_ uuid.UUID) (domain.Shipment, error) {
				return *pending, nil
			},
			SaveShipmentFn: func(_ *domain.Shipment) error {
				return fmt.Errorf("save failed")
			},
		}
		eventRepo := &MockEventRepository{}
		svc := newTestService(shipmentRepo, eventRepo)

		event, err := svc.AddStatusEvent(pending.GetID(), domain.StatusPickedUp)

		require.Error(t, err)
		assert.Nil(t, event)
		assert.Contains(t, err.Error(), "save failed")
	})

	t.Run("SaveEventError", func(t *testing.T) {
		pending := createPendingShipment()
		shipmentRepo := &MockShipmentRepository{
			GetShipmentByIDFn: func(_ uuid.UUID) (domain.Shipment, error) {
				return *pending, nil
			},
		}
		eventRepo := &MockEventRepository{
			SaveEventFn: func(_ domain.ShipmentEvent) error {
				return fmt.Errorf("event save failed")
			},
		}
		svc := newTestService(shipmentRepo, eventRepo)

		event, err := svc.AddStatusEvent(pending.GetID(), domain.StatusPickedUp)

		require.Error(t, err)
		assert.Nil(t, event)
		assert.Contains(t, err.Error(), "event save failed")
	})
}

func TestGetShipmentByID(t *testing.T) {
	t.Run("Found", func(t *testing.T) {
		s, _ := domain.NewShipment("REF-001", "NYC", "LAX")
		shipmentRepo := &MockShipmentRepository{
			GetShipmentByIDFn: func(_ uuid.UUID) (domain.Shipment, error) {
				return *s, nil
			},
		}
		eventRepo := &MockEventRepository{}
		svc := newTestService(shipmentRepo, eventRepo)

		result, err := svc.GetShipmentByID(s.GetID())

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, s.GetReferenceNumber(), result.GetReferenceNumber())
	})

	t.Run("NotFound", func(t *testing.T) {
		shipmentRepo := &MockShipmentRepository{}
		eventRepo := &MockEventRepository{}
		svc := newTestService(shipmentRepo, eventRepo)

		result, err := svc.GetShipmentByID(uuid.New())

		require.Error(t, err)
		assert.Nil(t, result)
		assert.True(t, errors.Is(err, domain.ErrShipmentNotFound))
	})
}

func TestGetShipmentHistory(t *testing.T) {
	shipmentID := uuid.New()

	t.Run("HasEvents", func(t *testing.T) {
		events := []domain.ShipmentEvent{
			{ShipmentID: shipmentID, Status: domain.StatusPending},
			{ShipmentID: shipmentID, Status: domain.StatusPickedUp},
		}
		shipmentRepo := &MockShipmentRepository{}
		eventRepo := &MockEventRepository{
			GetEventsByShipmentIDFn: func(_ uuid.UUID) ([]domain.ShipmentEvent, error) {
				return events, nil
			},
		}
		svc := newTestService(shipmentRepo, eventRepo)

		result, err := svc.GetShipmentHistory(shipmentID)

		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("NoEvents", func(t *testing.T) {
		shipmentRepo := &MockShipmentRepository{}
		eventRepo := &MockEventRepository{
			GetEventsByShipmentIDFn: func(_ uuid.UUID) ([]domain.ShipmentEvent, error) {
				return []domain.ShipmentEvent{}, nil
			},
		}
		svc := newTestService(shipmentRepo, eventRepo)

		result, err := svc.GetShipmentHistory(shipmentID)

		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Error", func(t *testing.T) {
		shipmentRepo := &MockShipmentRepository{}
		eventRepo := &MockEventRepository{
			GetEventsByShipmentIDFn: func(_ uuid.UUID) ([]domain.ShipmentEvent, error) {
				return nil, fmt.Errorf("event store error")
			},
		}
		svc := newTestService(shipmentRepo, eventRepo)

		result, err := svc.GetShipmentHistory(shipmentID)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "event store error")
	})
}
