package types

// AvailabilityQueryMessage defines a message where it is asked if a node is available to execute a computation
//
// - implements types.Message
// - implemented for project
type AvailabilityQueryMessage struct {
	// RequestID must be a unique identifier. Use xid.New().String() to generate
	// it.
	RequestID string // todo: isto pode ter problemas por se usar expanding ring search, se calhar avisar que é do mesmo nodo ou algo parecido
	// ou usar sempre o mesmo ID
	// Source is the address of the peer that sends the query
	Source string

	Budget uint // number of nodes needed for a computation

	AlreadyVisited map[string]bool // map indicating which nodes have already been visited, used to avoid sending spreading requests to already visited nodes
}

// AvailabilityResponseMessage defines a message a node send back to the originator of the query, indicating it is available for computation
//
// - implements types.Message
// - implemented for project
type AvailabilityResponseMessage struct {
	RequestID string
}

// ReservationCancellationMessage defines a message the node that requested the computation must send if it didn't manage
// to gather sufficient resources to perform the wanted computation
// ex: node A wants work divided in 3 nodes, but only gets 2 answers, being insufficient.
// Since these nodes are reserved, they must be told they can be available again
type ReservationCancellationMessage struct {
	RequestID string
}

// ComputationOrderMessage defines an order to execute the given executble with the given inputs
type ComputationOrderMessage struct {
	RequestID  string // to avoid hijacking other nodes
	Executable []byte
	Inputs     []string
}

// ComputationResultMessage defines the result of a remote computation
type ComputationResultMessage struct {
	RequestID string // to avoid hijacking other nodes
	Results   map[string]string
}

// TODO: o budged to availability Query é removido do guito de uma pessoa
// computation proposal e computation refusal + timeout
// todo: como evitar spam de requesições para roubar nodos: enviar o pedido em si tem um custo
// also: se calhar faz sentido cada nodo ter um custo base pela computação, para haver um tradeoff entre numero de nodos e speed
// TODO: a repsosta devia conter tb o buget desse nodo, de forma a previlegiar nodos com menos recursos
