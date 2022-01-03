package impl

import (
	"fmt"
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/types"
	"golang.org/x/xerrors"
)

// tries reserving this computer to execute a future computation
// if it is successful decrement the budget and add the node's IP to the already visited set
// if the budget is still bigger than 0, spread the query to non visited nodes
func (n *node) AvailabilityQueryMessageCallback(msg types.Message, _ transport.Packet) error {
	availabilityQueryMsg, ok := msg.(*types.AvailabilityQueryMessage)
	if !ok {
		return xerrors.Errorf("Conversion to AvailabilityQueryMessage failed")
	}

	managedToReserve, shouldBeForwarded := n.computationManager.tryReserve(availabilityQueryMsg.RequestID, availabilityQueryMsg.Budget)
	if managedToReserve {
		availabilityResponse := types.AvailabilityResponseMessage{
			RequestID: availabilityQueryMsg.RequestID,
		}
		availabilityResponseMsg, err := n.registry.MarshalMessage(availabilityResponse)
		if err != nil {
			return err
		}

		availabilityResponseHeader := transport.NewHeader(n.socket.GetAddress(), n.socket.GetAddress(), availabilityQueryMsg.Source, 0)
		availabilityResponsePkt := transport.Packet{
			Header: &availabilityResponseHeader,
			Msg:    &availabilityResponseMsg,
		}

		err = n.socket.Send(availabilityQueryMsg.Source, availabilityResponsePkt, 0)
		if err != nil {
			return err
		}
		fmt.Println("enviei resposta de volta")

		availabilityQueryMsg.Budget--
	} else {
		fmt.Println("problema, já está reservado")
	}

	if shouldBeForwarded {
		availabilityQueryMsg.AlreadyVisited[n.socket.GetAddress()] = true
		err := n.spreadAvailabilityQuery(*availabilityQueryMsg)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("não dei forward :o")
	}

	return nil
}

func (n *node) AvailabilityResponseMessageCallback(msg types.Message, pkt transport.Packet) error {
	availabilityResponseMsg, ok := msg.(*types.AvailabilityResponseMessage)
	if !ok {
		return xerrors.Errorf("Conversion to AvailabilityResponseMessage failed")
	}
	fmt.Println("received response:", availabilityResponseMsg, "from: ", pkt.Header.Source)
	n.computationManager.registerResponse(availabilityResponseMsg.RequestID, pkt.Header.Source)
	return nil
}

// send availability queries to neighbours that have not yet been visited
func (n *node) spreadAvailabilityQuery(msg types.AvailabilityQueryMessage) error {
	neighbourList := n.routingTable.getNeighboursList()
	var nonVisitedNeighbourList []string
	for _, address := range neighbourList {
		if _, alreadyVisited := msg.AlreadyVisited[address]; !alreadyVisited {
			nonVisitedNeighbourList = append(nonVisitedNeighbourList, address)
		}
	}

	if len(nonVisitedNeighbourList) == 0 {
		return nil
	}

	remainingBudget := msg.Budget
	if remainingBudget <= 0 {
		return nil
	}

	if int(remainingBudget) <= len(nonVisitedNeighbourList) {
		for _, neighbourAddress := range nonVisitedNeighbourList {
			pkt, err := n.adaptAvailabilityQueryPacket(msg, 1, neighbourAddress)
			if err != nil {
				return err
			}

			err = n.socket.Send(neighbourAddress, pkt, 0)
			if err != nil {
				return err
			}

			remainingBudget--
			if remainingBudget <= 0 {
				break
			}
		}
	} else {
		budgets := make([]uint, len(nonVisitedNeighbourList))
		i := 0
		for remainingBudget > 0 {
			i = i % len(neighbourList)
			budgets[i]++
			remainingBudget--
			i++
		}

		for i, neighbourAddress := range nonVisitedNeighbourList {
			pkt, err := n.adaptAvailabilityQueryPacket(msg, budgets[i], neighbourAddress)
			if err != nil {
				return err
			}

			err = n.socket.Send(neighbourAddress, pkt, 0)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// clones a message, adapting it's budget and wrapping it in a new header with an updated destination
func (n *node) adaptAvailabilityQueryPacket(msg types.AvailabilityQueryMessage, newBudget uint, peerAddress string) (transport.Packet, error) {
	availabilityRequest := types.AvailabilityQueryMessage{
		RequestID:      msg.RequestID,
		Source:         msg.Source,
		Budget:         newBudget,
		AlreadyVisited: msg.AlreadyVisited,
	}

	availabilityRequestMarshaled, err := n.registry.MarshalMessage(availabilityRequest)
	if err != nil {
		return transport.Packet{}, err
	}

	availabilityRequestHeader := transport.NewHeader(n.socket.GetAddress(), n.socket.GetAddress(), peerAddress, 0)
	return transport.Packet{
		Header: &availabilityRequestHeader,
		Msg:    &availabilityRequestMarshaled,
	}, nil
}
