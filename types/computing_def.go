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

	Budget uint
}

// AvailabilityResponseMessage defines a message a node send back to the originator of the query, indicating it is available for computation
//
// - implements types.Message
// - implemented for project
type AvailabilityResponseMessage struct {
	RequestID string
	// how much the node has in their balance
	CurrentBalance int //TODO: this can maybe be derived from the blockchain
}

// TODO: não enviar o pedido para a source
// TODO: os nodos devem ter um mapa de requests a que já responderam, para não responder duplicadamente
// TODO: what happens if the whole network alreayd received the requests/or is occupied? is the packet just dropped? USe ttl?
//TODO: sending budget can be very large, only remove as many as how many responses you got
// TODO: o budged to availability Query é removido do guito de uma pessoa
// computation proposal e computation refusal + timeout
// todo: como evitar spam de requesições para roubar nodos: enviar o pedido em si tem um custo
// also: se calhar faz sentido cada nodo ter um custo base pela computação, para haver um tradeoff entre numero de nodos e speed
// TODO: a repsosta devia conter tb o buget desse nodo, de forma a previlegiar nodos com menos recursos
