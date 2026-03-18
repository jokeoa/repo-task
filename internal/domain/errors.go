package domain

import "errors"

var (
	ErrInvalidTransition = errors.New("invalid status transition")
	// ErrMissingField        = errors.New("missing required field")
	ErrInvalidShipmentData = errors.New("invalid shipment data")
)
