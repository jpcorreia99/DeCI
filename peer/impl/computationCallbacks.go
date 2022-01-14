package impl

import (
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/types"
	"golang.org/x/xerrors"
	"os/exec"
	"strings"
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

		availabilityQueryMsg.Budget--
	}

	if shouldBeForwarded {
		availabilityQueryMsg.AlreadyVisited[n.socket.GetAddress()] = true
		err := n.spreadAvailabilityQuery(*availabilityQueryMsg)
		if err != nil {
			return err
		}
	}

	return nil
}

func (n *node) AvailabilityResponseMessageCallback(msg types.Message, pkt transport.Packet) error {
	availabilityResponseMsg, ok := msg.(*types.AvailabilityResponseMessage)
	if !ok {
		return xerrors.Errorf("Conversion to AvailabilityResponseMessage failed")
	}

	n.computationManager.registerAvailability(availabilityResponseMsg.RequestID, pkt.Header.Source)
	return nil
}

func (n *node) ReservationCancellationMessageCallback(msg types.Message, _ transport.Packet) error {
	reservationCancellationMsg, ok := msg.(*types.ReservationCancellationMessage)
	if !ok {
		return xerrors.Errorf("Conversion to ReservationCancellationMessage failed")
	}

	n.computationManager.cancelReservation(reservationCancellationMsg.RequestID)

	return nil
}

func (n *node) ComputationOrderMessageCallback(msg types.Message, pkt transport.Packet) error {
	computationOrderMsg, ok := msg.(*types.ComputationOrderMessage)
	if !ok {
		return xerrors.Errorf("Conversion to ComputationOrderMessage failed")
	}

	requestID := computationOrderMsg.RequestID
	idWasPromised := n.computationManager.signalExecutableReceived(requestID)
	if !idWasPromised {
		return nil
	}

	executable := computationOrderMsg.Executable
	inputs := computationOrderMsg.Inputs

	filename, err := saveExecutable(executable, computationOrderMsg.FileExtension)
	if err != nil {
		return err
	}

	answerMap := make(map[string]string)
	for _, line := range inputs {
		codeArgs := strings.Split(line, ",")
		app := computationOrderMsg.ExecutionArgs[0]
		args := make([]string, 0, 1+len(computationOrderMsg.ExecutionArgs)-1+len(codeArgs))
		args = append(args, computationOrderMsg.ExecutionArgs[1:]...)
		args = append(args, filename)
		args = append(args, codeArgs...)
		output, err := exec.Command(app, args...).Output()
		if err != nil {
			return err
		}
		answerMap[line] = string(output)
	}

	pktToSend, err := n.createComputationResultPacket(requestID, answerMap, pkt.Header.Source)
	if err != nil {
		return err
	}

	err = n.socket.Send(pkt.Header.Source, pktToSend, 0)
	n.computationManager.cancelReservation(requestID)
	return err
}

func (n *node) ComputationResultMessageCallback(msg types.Message, _ transport.Packet) error {
	computationResultMsg, ok := msg.(*types.ComputationResultMessage)
	if !ok {
		return xerrors.Errorf("Conversion to ComputationResultMessage failed")
	}

	n.computationManager.storeResults(computationResultMsg.RequestID, computationResultMsg.Results)
	return nil
}

// send availability queries to the neighbours that have not yet been visited
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

func (n *node) createComputationResultPacket(requestID string, results map[string]string, peerAddress string) (transport.Packet, error) {
	computationResult := types.ComputationResultMessage{
		RequestID: requestID,
		Results:   results,
	}

	computationResultMarshaled, err := n.registry.MarshalMessage(computationResult)
	if err != nil {
		return transport.Packet{}, err
	}

	computationResultHeader := transport.NewHeader(n.socket.GetAddress(), n.socket.GetAddress(), peerAddress, 0)
	return transport.Packet{
		Header: &computationResultHeader,
		Msg:    &computationResultMarshaled,
	}, nil

}
