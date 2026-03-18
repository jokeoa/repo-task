package domain

import "errors"

var (
	ErrInvalidTransition   = errors.New("invalid status transition")
	ErrInvalidShipmentData = errors.New("invalid shipment data")
	ErrShipmentNotFound    = errors.New("shipment not found")
	ErrInvalidStatus       = errors.New("invalid status")
)
