package types

import (
	"fmt"
)

// -----------------------------------------------------------------------------
// JoinMessage

// NewEmpty implements types.Message.
func (c JoinMessage) NewEmpty() Message {
	return &JoinMessage{}
}

// Name implements types.Message.
func (JoinMessage) Name() string {
	return "join"
}

// String implements types.Message.
func (c JoinMessage) String() string {
	return fmt.Sprintf("<%s>", c.Message)
}

// HTML implements types.Message.
func (c JoinMessage) HTML() string {
	return c.String()
}

// -----------------------------------------------------------------------------
// AckJoinMessage

// NewEmpty implements types.Message.
func (c AckJoinMessage) NewEmpty() Message {
	return &AckJoinMessage{}
}

// Name implements types.Message.
func (AckJoinMessage) Name() string {
	return "ackjoin"
}

// String implements types.Message.
func (c AckJoinMessage) String() string {
	return fmt.Sprintf("<%s>", c.Message)
}

// HTML implements types.Message.
func (c AckJoinMessage) HTML() string {
	return c.String()
}
