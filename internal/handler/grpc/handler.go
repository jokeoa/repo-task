package grpchandler

import (
	"context"
	"errors"

	"tracker-task/internal/domain"
	"tracker-task/internal/usecase"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "tracker-task/gen/proto"
)

type ShipmentHandler struct {
	pb.UnimplementedShipmentServiceServer
	service *usecase.ShipmentService
}

func NewShipmentHandler(service *usecase.ShipmentService) *ShipmentHandler {
	return &ShipmentHandler{service: service}
}

func (h *ShipmentHandler) CreateShipment(_ context.Context, req *pb.CreateShipmentRequest) (*pb.ShipmentResponse, error) {
	units, err := protoUnitsToDomain(req.GetUnits())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid units: %v", err)
	}

	input := usecase.CreateShipmentInput{
		Reference:     req.GetReferenceNumber(),
		Origin:        req.GetOrigin(),
		Destination:   req.GetDestination(),
		Units:         units,
		Amount:        domain.Money(req.GetAmountCents()),
		DriverRevenue: domain.Money(req.GetDriverRevenueCents()),
	}

	if req.GetDriver() != "" {
		driverID, err := uuid.Parse(req.GetDriver())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid driver ID: %v", err)
		}
		input.Driver = &driverID
	}

	shipment, err := h.service.CreateShipment(input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidShipmentData) {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return shipmentToProto(shipment), nil
}

func (h *ShipmentHandler) AddStatusEvent(_ context.Context, req *pb.AddStatusEventRequest) (*pb.ShipmentEventResponse, error) {
	shipmentID, err := uuid.Parse(req.GetShipmentId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid shipment ID: %v", err)
	}

	parsedStatus, err := domain.StatusFromString(req.GetStatus())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	event, err := h.service.AddStatusEvent(shipmentID, parsedStatus)
	if err != nil {
		if errors.Is(err, domain.ErrShipmentNotFound) {
			return nil, status.Errorf(codes.NotFound, "%v", err)
		}
		if errors.Is(err, domain.ErrInvalidTransition) {
			return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return eventToProto(*event), nil
}

func (h *ShipmentHandler) GetShipmentByID(_ context.Context, req *pb.GetShipmentRequest) (*pb.ShipmentResponse, error) {
	shipmentID, err := uuid.Parse(req.GetShipmentId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid shipment ID: %v", err)
	}

	shipment, err := h.service.GetShipmentByID(shipmentID)
	if err != nil {
		if errors.Is(err, domain.ErrShipmentNotFound) {
			return nil, status.Errorf(codes.NotFound, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return shipmentToProto(shipment), nil
}

func (h *ShipmentHandler) GetShipmentHistory(_ context.Context, req *pb.GetShipmentRequest) (*pb.ShipmentEventHistoryResponse, error) {
	shipmentID, err := uuid.Parse(req.GetShipmentId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid shipment ID: %v", err)
	}

	events, err := h.service.GetShipmentHistory(shipmentID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "%v", err)
	}

	return &pb.ShipmentEventHistoryResponse{
		ShipmentId: shipmentID.String(),
		Events:     eventsToProto(events),
	}, nil
}
