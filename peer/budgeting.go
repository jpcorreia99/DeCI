package peer

// Budgeting describes functions to manage a peers budget using a blockchain.
type Budgeting interface {
	// UpdateBudget updates the budgets of the peers participating in a computation using consensus. First the peer that
	// initiated the computation, decreases its budget by the total cost and announces it to the rest of the network.
	// Then, it proposes a value which is the total cost of the computation divided by the number of nodes that
	// participated in the computation, along with the list of peers that participated in the computation.

	UpdateBudget(computationID string, participantAmountMap map[string]float64) error
}
