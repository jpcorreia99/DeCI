package types

import "fmt"

// -----------------------------------------------------------------------------
// AvailabilityQueryMessage

// NewEmpty implements types.Message.
func (a AvailabilityQueryMessage) NewEmpty() Message {
	return &AvailabilityQueryMessage{}
}

// Name implements types.Message.
func (a AvailabilityQueryMessage) Name() string {
	return "availabilityquery"
}

// String implements types.Message.
func (a AvailabilityQueryMessage) String() string {
	return fmt.Sprintf("{availabilityquery %s - %s - %d}", a.RequestID, a.Source, a.Budget)
}

// HTML implements types.Message.
func (a AvailabilityQueryMessage) HTML() string {
	return a.String()
}

// -----------------------------------------------------------------------------
// AvailabilityResponseMessage

// NewEmpty implements types.Message.
func (a AvailabilityResponseMessage) NewEmpty() Message {
	return &AvailabilityResponseMessage{}
}

// Name implements types.Message.
func (a AvailabilityResponseMessage) Name() string {
	return "availabilityresponse"
}

// String implements types.Message.
func (a AvailabilityResponseMessage) String() string {
	return fmt.Sprintf("{availabilityresponse %s}", a.RequestID)
}

// HTML implements types.Message.
func (a AvailabilityResponseMessage) HTML() string {
	return a.String()
}

// -----------------------------------------------------------------------------
// ReservationCancellationMessage

// NewEmpty implements types.Message.
func (r ReservationCancellationMessage) NewEmpty() Message {
	return &ReservationCancellationMessage{}
}

// Name implements types.Message.
func (r ReservationCancellationMessage) Name() string {
	return "reservationcancellation"
}

// String implements types.Message.
func (r ReservationCancellationMessage) String() string {
	return fmt.Sprintf("{reservationcancellation %d}", r.RequestID)
}

// HTML implements types.Message.
func (r ReservationCancellationMessage) HTML() string {
	return r.String()
}

// -----------------------------------------------------------------------------
// ComputationOrderMessage

// NewEmpty implements types.Message.
func (c ComputationOrderMessage) NewEmpty() Message {
	return &ComputationOrderMessage{}
}

// Name implements types.Message.
func (c ComputationOrderMessage) Name() string {
	return "computationorder"
}

// String implements types.Message.
func (c ComputationOrderMessage) String() string {
	return fmt.Sprintf("{computationorder %s}", c.RequestID)
}

// HTML implements types.Message.
func (c ComputationOrderMessage) HTML() string {
	return c.String()
}

// -----------------------------------------------------------------------------
// ComputationResultMessage

// NewEmpty implements types.Message.
func (c ComputationResultMessage) NewEmpty() Message {
	return &ComputationResultMessage{}
}

// Name implements types.Message.
func (c ComputationResultMessage) Name() string {
	return "computationresult"
}

// String implements types.Message.
func (c ComputationResultMessage) String() string {
	return fmt.Sprintf("{computationresult %s}", c.RequestID)
}

// HTML implements types.Message.
func (c ComputationResultMessage) HTML() string {
	return c.String()
}
