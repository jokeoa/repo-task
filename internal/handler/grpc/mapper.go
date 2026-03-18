package grpchandler

import (
	"tracker-task/internal/domain"

	"github.com/google/uuid"

	pb "tracker-task/gen/proto"
)

func shipmentToProto(s *domain.Shipment) *pb.ShipmentResponse {
	statusStr, _ := s.GetCurrentStatus().GetStatus()

	units := make([]*pb.Unit, len(s.GetUnits()))
	for i, u := range s.GetUnits() {
		units[i] = &pb.Unit{
			Id:          u.ID.String(),
			Description: u.Description,
		}
	}

	resp := &pb.ShipmentResponse{
		ShipmentId:      s.GetID().String(),
		ReferenceNumber: s.GetReferenceNumber(),
		Origin:          s.GetOrigin(),
		Destination:     s.GetDestination(),
		Status:          statusStr,
		Units:           units,
		AmountCents:     int64(s.GetAmount()),
		DriverRevenueCents: int64(s.GetDriverRevenue()),
	}

	if d := s.GetDriver(); d != nil {
		resp.Driver = d.String()
	}

	return resp
}

func eventToProto(e domain.ShipmentEvent) *pb.ShipmentEventResponse {
	statusStr, _ := e.Status.GetStatus()
	return &pb.ShipmentEventResponse{
		ShipmentId: e.ShipmentID.String(),
		Status:     statusStr,
		Timestamp:  e.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func eventsToProto(events []domain.ShipmentEvent) []*pb.ShipmentEventResponse {
	result := make([]*pb.ShipmentEventResponse, len(events))
	for i, e := range events {
		result[i] = eventToProto(e)
	}
	return result
}

func protoUnitsToDomain(units []*pb.Unit) ([]domain.Unit, error) {
	result := make([]domain.Unit, len(units))
	for i, u := range units {
		id, err := uuid.Parse(u.GetId())
		if err != nil {
			id = uuid.New()
		}
		result[i] = domain.Unit{
			ID:          id,
			Description: u.GetDescription(),
		}
	}
	return result, nil
}
