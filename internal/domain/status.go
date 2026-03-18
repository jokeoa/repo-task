package domain

import "fmt"

type ShipmentStatus int

const (
	StatusUnknown ShipmentStatus = iota
	StatusPending
	StatusPickedUp
	StatusInTransit
	StatusDelivered
	StatusCancelled
)

func (ss ShipmentStatus) GetStatus() (string, error) {
	switch ss {
	case StatusPending:
		return "Pending", nil
	case StatusPickedUp:
		return "Picked Up", nil
	case StatusInTransit:
		return "In Transit", nil
	case StatusDelivered:
		return "Delivered", nil
	case StatusCancelled:
		return "Cancelled", nil
	case StatusUnknown:
		return "Unknown", nil
	default:
		return "", fmt.Errorf("invalid shipment status: %d", ss)
	}
}

func (ss ShipmentStatus) CanTransitionTo(next ShipmentStatus) bool {
	switch ss {
	case StatusPending:
		return next == StatusPickedUp || next == StatusCancelled
	case StatusPickedUp:
		return next == StatusInTransit || next == StatusCancelled
	case StatusInTransit:
		return next == StatusDelivered || next == StatusCancelled
	case StatusDelivered:
		return false
	case StatusCancelled:
		return false
	default:
		return false
	}
}
