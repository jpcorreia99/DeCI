package impl

import (
	"fmt"
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/types"
	"golang.org/x/xerrors"
)

func (n *node) AvailabilityQueryMessageCallback(msg types.Message, _ transport.Packet) error {
	availabilityQueryMsg, ok := msg.(*types.AvailabilityQueryMessage)
	if !ok {
		return xerrors.Errorf("Conversion to AvailabilityQueryMessage failed")
	}

	managedToReserve, currentBalance := n.computationManager.tryReserve(availabilityQueryMsg.RequestID)
	if managedToReserve {
		availabilityResponse := types.AvailabilityResponseMessage{
			RequestID:      availabilityQueryMsg.RequestID,
			CurrentBalance: currentBalance,
		}
		availabilityResponseMsg, err := n.registry.MarshalMessage(availabilityResponse)
		if err != nil {
			return err
		}

		availabilityResponseHeader := transport.NewHeader(n.socket.GetAddress(), n.socket.GetAddress(), availabilityQueryMsg.Source, 0)
		availabilityResponsPkt := transport.Packet{
			Header: &availabilityResponseHeader,
			Msg:    &availabilityResponseMsg,
		}

		err = n.socket.Send(availabilityQueryMsg.Source, availabilityResponsPkt, 0)
		if err != nil {
			return err
		}
		fmt.Println("enviei resposta de volta")
	}
	fmt.Println("problema, já está reservado")
	return nil
}

func (n *node) AvailabilityResponseMessageCallback(msg types.Message, pkt transport.Packet) error {
	availabilityResponseMsg, ok := msg.(*types.AvailabilityResponseMessage)
	if !ok {
		return xerrors.Errorf("Conversion to AvailabilityResponseMessage failed")
	}
	fmt.Println("received response:", availabilityResponseMsg, "from: ", pkt.Header.Source)
	return nil
}
