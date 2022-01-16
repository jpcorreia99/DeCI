package impl

import (
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/types"
	"golang.org/x/xerrors"
)

func (n *node) JoinMessageCallback(msg types.Message, pkt transport.Packet) error {
	joinMsg, ok := msg.(*types.JoinMessage)
	if !ok {
		return xerrors.Errorf("Conversion to Chat Message failed")
	}

	joiningNodeAddress := joinMsg.Message

	//fmt.Printf("At %v, received join message from %v \n", n.socket.GetAddress(), joiningNodeAddress)

	if n.peerMap.contains(joiningNodeAddress) {
		return nil
	}

	n.configuration.TotalPeers.Increase()

	n.peerMap.add(joiningNodeAddress)

	n.AddPeer(joiningNodeAddress)

	ackJoinMsg := types.AckJoinMessage{
		Message: n.socket.GetAddress(),
	}

	ackJoinMsgMarshaled, err := n.registry.MarshalMessage(ackJoinMsg)

	if err != nil {
		return err
	}

	header := transport.NewHeader(n.socket.GetAddress(), n.socket.GetAddress(), joiningNodeAddress, 1)
	newPkt := transport.Packet{
		Header: &header,
		Msg:    &ackJoinMsgMarshaled,
	}

	err = n.socket.Send(joiningNodeAddress, newPkt, 0)

	if err != nil {
		return err
	}

	return nil
}

func (n *node) AckJoinMessageCallback(msg types.Message, pkt transport.Packet) error {
	ackJoinMsg, ok := msg.(*types.AckJoinMessage)
	if !ok {
		return xerrors.Errorf("Conversion to Chat Message failed")
	}

	ackJoiningNodeAddress := ackJoinMsg.Message
	//fmt.Printf("At %v, received ACK join message from %v \n", n.socket.GetAddress(), ackJoiningNodeAddress)

	if !n.joinAckMap.contains(ackJoiningNodeAddress) {
		n.joinAckMap.add(ackJoiningNodeAddress)
	} else {
		println("Already received ack from this peer")
	}

	//fmt.Printf("New ack map size : %v and total peers %v\n", n.joinAckMap.getSize(), n.configuration.TotalPeers.Get())

	if n.joinAckMap.getSize() > int(n.configuration.TotalPeers.Get()) {
		println("Updating number of total peers!")
		n.configuration.TotalPeers.Set(uint(n.joinAckMap.getSize()))
	}

	if n.joinAckMap.getSize() > int(n.configuration.TotalPeers.Get())/2 && !n.hasJoined.get() {
		println("Joining the network!")
		n.hasJoined.set(true)
	}

	return nil
}
