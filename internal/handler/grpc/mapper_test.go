package grpchandler

import (
	"testing"
	"time"

	"tracker-task/internal/domain"

	pb "tracker-task/gen/proto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShipmentToProto(t *testing.T) {
	s, err := domain.NewShipment("REF-001", "NYC", "LAX")
	require.NoError(t, err)

	s.AddUnit(domain.Unit{ID: uuid.New(), Description: "Box A"})
	s.SetAmount(5000)
	s.SetDriverRevenue(1500)

	driverID := uuid.New()
	s.SetDriver(driverID)

	s.AddEvent(domain.StatusPending)

	resp := shipmentToProto(s)

	assert.Equal(t, s.GetID().String(), resp.GetShipmentId())
	assert.Equal(t, "REF-001", resp.GetReferenceNumber())
	assert.Equal(t, "NYC", resp.GetOrigin())
	assert.Equal(t, "LAX", resp.GetDestination())
	assert.Equal(t, "Pending", resp.GetStatus())
	assert.Len(t, resp.GetUnits(), 1)
	assert.Equal(t, "Box A", resp.GetUnits()[0].GetDescription())
	assert.Equal(t, driverID.String(), resp.GetDriver())
	assert.Equal(t, int64(5000), resp.GetAmountCents())
	assert.Equal(t, int64(1500), resp.GetDriverRevenueCents())
}

func TestShipmentToProto_NoDriver(t *testing.T) {
	s, err := domain.NewShipment("REF-002", "NYC", "LAX")
	require.NoError(t, err)
	s.AddEvent(domain.StatusPending)

	resp := shipmentToProto(s)

	assert.Empty(t, resp.GetDriver())
}

func TestEventToProto(t *testing.T) {
	now := time.Now()
	e := domain.ShipmentEvent{
		ShipmentID: uuid.New(),
		Status:     domain.StatusPickedUp,
		CreatedAt:  now,
	}

	resp := eventToProto(e)

	assert.Equal(t, e.ShipmentID.String(), resp.GetShipmentId())
	assert.Equal(t, "Picked Up", resp.GetStatus())
	assert.Equal(t, now.Format("2006-01-02T15:04:05Z07:00"), resp.GetTimestamp())
}

func TestEventsToProto(t *testing.T) {
	shipmentID := uuid.New()
	events := []domain.ShipmentEvent{
		{ShipmentID: shipmentID, Status: domain.StatusPending, CreatedAt: time.Now()},
		{ShipmentID: shipmentID, Status: domain.StatusPickedUp, CreatedAt: time.Now()},
	}

	result := eventsToProto(events)

	assert.Len(t, result, 2)
	assert.Equal(t, "Pending", result[0].GetStatus())
	assert.Equal(t, "Picked Up", result[1].GetStatus())
}

func TestEventsToProto_Empty(t *testing.T) {
	result := eventsToProto([]domain.ShipmentEvent{})
	assert.Empty(t, result)
}

func TestProtoUnitsToDomain(t *testing.T) {
	unitID := uuid.New()
	protoUnits := []*pb.Unit{
		{Id: unitID.String(), Description: "Box A"},
		{Id: "invalid-uuid", Description: "Box B"},
	}

	units, err := protoUnitsToDomain(protoUnits)

	require.NoError(t, err)
	require.Len(t, units, 2)
	assert.Equal(t, unitID, units[0].ID)
	assert.Equal(t, "Box A", units[0].Description)
	assert.NotEqual(t, uuid.Nil, units[1].ID)
	assert.Equal(t, "Box B", units[1].Description)
}

func TestProtoUnitsToDomain_Empty(t *testing.T) {
	units, err := protoUnitsToDomain([]*pb.Unit{})
	require.NoError(t, err)
	assert.Empty(t, units)
}
