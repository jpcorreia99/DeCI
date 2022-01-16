package impl

import (
	"go.dedis.ch/cs438/types"
	"time"
)

func (n *node) JoinNetwork(knownNeighbourAddress string) error {
	joinMsg := types.JoinMessage{
		Message: n.socket.GetAddress(),
	}

	n.AddPeer(knownNeighbourAddress)

	joinMsgMarshaled, err := n.registry.MarshalMessage(joinMsg)

	if err != nil {
		return err
	}

	err = n.Broadcast(joinMsgMarshaled)

	if err != nil {
		return err
	}

	time.Sleep(time.Second * 5)

	for !n.hasJoined.get() {

	}

	return nil
}

func (n *node) GetAddress() string {
	return n.socket.GetAddress()
}

func (n *node) GetBudget() float64 {
	return n.localBudget
}
