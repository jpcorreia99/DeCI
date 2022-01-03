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
	return fmt.Sprintf("{availabilityquery %d - %d - %d}", a.RequestID, a.Source, a.Budget)
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
	return fmt.Sprintf("{availabilityresponse %d}", a.RequestID)
}

// HTML implements types.Message.
func (a AvailabilityResponseMessage) HTML() string {
	return a.String()
}
