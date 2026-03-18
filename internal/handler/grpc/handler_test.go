package grpchandler_test

import (
	"context"
	"testing"

	grpchandler "tracker-task/internal/handler/grpc"
	"tracker-task/internal/infra/memory"
	"tracker-task/internal/usecase"

	pb "tracker-task/gen/proto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func setupHandler() *grpchandler.ShipmentHandler {
	shipmentRepo := memory.NewShipmentRepository()
	eventRepo := memory.NewEventRepository()
	svc := usecase.NewShipmentService(shipmentRepo, eventRepo)
	return grpchandler.NewShipmentHandler(svc)
}

func createTestShipment(t *testing.T, handler *grpchandler.ShipmentHandler) *pb.ShipmentResponse {
	t.Helper()
	resp, err := handler.CreateShipment(context.Background(), &pb.CreateShipmentRequest{
		ReferenceNumber: "REF-001",
		Origin:          "NYC",
		Destination:     "LAX",
		Units: []*pb.Unit{
			{Id: uuid.New().String(), Description: "Box A"},
		},
	})
	require.NoError(t, err)
	return resp
}

func TestHandler_CreateShipment(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		handler := setupHandler()
		unitID := uuid.New().String()

		resp, err := handler.CreateShipment(context.Background(), &pb.CreateShipmentRequest{
			ReferenceNumber: "REF-001",
			Origin:          "NYC",
			Destination:     "LAX",
			Units: []*pb.Unit{
				{Id: unitID, Description: "Box A"},
			},
			AmountCents:       5000,
			DriverRevenueCents: 1500,
		})

		require.NoError(t, err)
		assert.Equal(t, "REF-001", resp.GetReferenceNumber())
		assert.Equal(t, "NYC", resp.GetOrigin())
		assert.Equal(t, "LAX", resp.GetDestination())
		assert.Equal(t, "Pending", resp.GetStatus())
		assert.Len(t, resp.GetUnits(), 1)
		assert.Equal(t, int64(5000), resp.GetAmountCents())
		assert.Equal(t, int64(1500), resp.GetDriverRevenueCents())
	})

	t.Run("WithDriver", func(t *testing.T) {
		handler := setupHandler()
		driverID := uuid.New().String()

		resp, err := handler.CreateShipment(context.Background(), &pb.CreateShipmentRequest{
			ReferenceNumber: "REF-002",
			Origin:          "NYC",
			Destination:     "LAX",
			Units:           []*pb.Unit{{Id: uuid.New().String(), Description: "Box"}},
			Driver:          driverID,
		})

		require.NoError(t, err)
		assert.Equal(t, driverID, resp.GetDriver())
	})

	t.Run("InvalidDriver", func(t *testing.T) {
		handler := setupHandler()

		_, err := handler.CreateShipment(context.Background(), &pb.CreateShipmentRequest{
			ReferenceNumber: "REF-003",
			Origin:          "NYC",
			Destination:     "LAX",
			Units:           []*pb.Unit{{Id: uuid.New().String(), Description: "Box"}},
			Driver:          "not-a-uuid",
		})

		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("EmptyUnits", func(t *testing.T) {
		handler := setupHandler()

		_, err := handler.CreateShipment(context.Background(), &pb.CreateShipmentRequest{
			ReferenceNumber: "REF-004",
			Origin:          "NYC",
			Destination:     "LAX",
		})

		require.Error(t, err)
		assert.Equal(t, codes.Internal, status.Code(err))
	})

	t.Run("EmptyRef", func(t *testing.T) {
		handler := setupHandler()

		_, err := handler.CreateShipment(context.Background(), &pb.CreateShipmentRequest{
			ReferenceNumber: "",
			Origin:          "NYC",
			Destination:     "LAX",
			Units:           []*pb.Unit{{Id: uuid.New().String(), Description: "Box"}},
		})

		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}

func TestHandler_AddStatusEvent(t *testing.T) {
	t.Run("ValidTransition", func(t *testing.T) {
		handler := setupHandler()
		created := createTestShipment(t, handler)

		resp, err := handler.AddStatusEvent(context.Background(), &pb.AddStatusEventRequest{
			ShipmentId: created.GetShipmentId(),
			Status:     "Picked Up",
		})

		require.NoError(t, err)
		assert.Equal(t, "Picked Up", resp.GetStatus())
		assert.Equal(t, created.GetShipmentId(), resp.GetShipmentId())
		assert.NotEmpty(t, resp.GetTimestamp())
	})

	t.Run("InvalidTransition", func(t *testing.T) {
		handler := setupHandler()
		created := createTestShipment(t, handler)

		_, err := handler.AddStatusEvent(context.Background(), &pb.AddStatusEventRequest{
			ShipmentId: created.GetShipmentId(),
			Status:     "Delivered",
		})

		require.Error(t, err)
		assert.Equal(t, codes.FailedPrecondition, status.Code(err))
	})

	t.Run("InvalidShipmentID", func(t *testing.T) {
		handler := setupHandler()

		_, err := handler.AddStatusEvent(context.Background(), &pb.AddStatusEventRequest{
			ShipmentId: "not-a-uuid",
			Status:     "Picked Up",
		})

		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	t.Run("ShipmentNotFound", func(t *testing.T) {
		handler := setupHandler()

		_, err := handler.AddStatusEvent(context.Background(), &pb.AddStatusEventRequest{
			ShipmentId: uuid.New().String(),
			Status:     "Picked Up",
		})

		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("InvalidStatus", func(t *testing.T) {
		handler := setupHandler()
		created := createTestShipment(t, handler)

		_, err := handler.AddStatusEvent(context.Background(), &pb.AddStatusEventRequest{
			ShipmentId: created.GetShipmentId(),
			Status:     "Lost",
		})

		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}

func TestHandler_GetShipmentByID(t *testing.T) {
	t.Run("Found", func(t *testing.T) {
		handler := setupHandler()
		created := createTestShipment(t, handler)

		resp, err := handler.GetShipmentByID(context.Background(), &pb.GetShipmentRequest{
			ShipmentId: created.GetShipmentId(),
		})

		require.NoError(t, err)
		assert.Equal(t, created.GetShipmentId(), resp.GetShipmentId())
		assert.Equal(t, "REF-001", resp.GetReferenceNumber())
	})

	t.Run("NotFound", func(t *testing.T) {
		handler := setupHandler()

		_, err := handler.GetShipmentByID(context.Background(), &pb.GetShipmentRequest{
			ShipmentId: uuid.New().String(),
		})

		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	t.Run("InvalidID", func(t *testing.T) {
		handler := setupHandler()

		_, err := handler.GetShipmentByID(context.Background(), &pb.GetShipmentRequest{
			ShipmentId: "not-a-uuid",
		})

		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}

func TestHandler_GetShipmentHistory(t *testing.T) {
	t.Run("HasEvents", func(t *testing.T) {
		handler := setupHandler()
		created := createTestShipment(t, handler)

		handler.AddStatusEvent(context.Background(), &pb.AddStatusEventRequest{
			ShipmentId: created.GetShipmentId(),
			Status:     "Picked Up",
		})

		resp, err := handler.GetShipmentHistory(context.Background(), &pb.GetShipmentRequest{
			ShipmentId: created.GetShipmentId(),
		})

		require.NoError(t, err)
		assert.Equal(t, created.GetShipmentId(), resp.GetShipmentId())
		assert.Len(t, resp.GetEvents(), 2) // Pending + Picked Up
	})

	t.Run("NoEvents", func(t *testing.T) {
		handler := setupHandler()

		resp, err := handler.GetShipmentHistory(context.Background(), &pb.GetShipmentRequest{
			ShipmentId: uuid.New().String(),
		})

		require.NoError(t, err)
		assert.Empty(t, resp.GetEvents())
	})

	t.Run("InvalidID", func(t *testing.T) {
		handler := setupHandler()

		_, err := handler.GetShipmentHistory(context.Background(), &pb.GetShipmentRequest{
			ShipmentId: "not-a-uuid",
		})

		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}

func TestHandler_FullLifecycle(t *testing.T) {
	handler := setupHandler()

	// Create
	created, err := handler.CreateShipment(context.Background(), &pb.CreateShipmentRequest{
		ReferenceNumber: "REF-LIFE",
		Origin:          "NYC",
		Destination:     "LAX",
		Units:           []*pb.Unit{{Id: uuid.New().String(), Description: "Box"}},
		AmountCents:     10000,
	})
	require.NoError(t, err)
	assert.Equal(t, "Pending", created.GetStatus())

	shipmentID := created.GetShipmentId()

	// Advance through lifecycle
	transitions := []string{"Picked Up", "In Transit", "Delivered"}
	for _, s := range transitions {
		_, err := handler.AddStatusEvent(context.Background(), &pb.AddStatusEventRequest{
			ShipmentId: shipmentID,
			Status:     s,
		})
		require.NoError(t, err)
	}

	// Verify final status
	final, err := handler.GetShipmentByID(context.Background(), &pb.GetShipmentRequest{
		ShipmentId: shipmentID,
	})
	require.NoError(t, err)
	assert.Equal(t, "Delivered", final.GetStatus())

	// Verify full history
	history, err := handler.GetShipmentHistory(context.Background(), &pb.GetShipmentRequest{
		ShipmentId: shipmentID,
	})
	require.NoError(t, err)
	assert.Len(t, history.GetEvents(), 4) // Pending + 3 transitions

	expectedStatuses := []string{"Pending", "Picked Up", "In Transit", "Delivered"}
	for i, evt := range history.GetEvents() {
		assert.Equal(t, expectedStatuses[i], evt.GetStatus())
	}
}
