package types

// JoinMessage is a message to indicate the a new node in joining the network.
//
// - implements types.Message
type JoinMessage struct {
	Message string
}

// AckJoinMessage is a message to indicate the a new node in joining the network.
//
// - implements types.Message
type AckJoinMessage struct {
	Message string
}
