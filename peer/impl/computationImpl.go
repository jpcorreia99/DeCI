package impl

import (
	"fmt"
	"github.com/rs/xid"
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/types"
	"golang.org/x/xerrors"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var EXECUTABLE_LOCATION = "../executables/"

func (n *node) Compute(executable []byte, executionArgs []string, fileExtension string, inputs []byte, numberOfRequestedNodes uint) ([]byte, error) {
	if numberOfRequestedNodes > n.configuration.TotalPeers.Get()-1 {
		return nil, xerrors.Errorf("Number of nodes to request above total number of nodes: %v - %v ",
			numberOfRequestedNodes, n.configuration.TotalPeers.Get()-1)
	}

	requestID := xid.New().String()

	// --------- STEP ----------
	// save the executable, split the input data and estimate the cost of computing per unit
	// also register on the computation manager the first resuls and retrieve the channel to

	//start := time.Now()

	filename, err := saveExecutable(executable, fileExtension)
	if err != nil {
		return nil, err
	}

	inputData := splitData(inputs)
	costPerUnit, alreadyCalculatedResults, err := estimateCost(filename, executionArgs, inputData)
	if err != nil {
		return nil, err
	}

	totalCost := costPerUnit*float64(len(inputData)-len(alreadyCalculatedResults)) + float64(numberOfRequestedNodes)
	if totalCost > n.localBudget {
		return nil, xerrors.Errorf("Total operation cost above available budget %v - %v", totalCost, n.localBudget)
	}

	results := make(map[string]string)
	for input, output := range alreadyCalculatedResults {
		results[input] = output
	}

	computationConclusionChannel := n.computationManager.registerComputation(requestID, uint(len(inputData)))
	n.computationManager.storeResults(requestID, results)

	//elapsed := time.Since(start)
	//fmt.Println("Cost estimation duration: ", elapsed.Seconds())
	// remove from the inputs the ones already calculated
	remainingInputs := make([]string, 0, len(inputData)-len(alreadyCalculatedResults))
	for _, val := range inputData {
		if _, isAlreadyCalculated := alreadyCalculatedResults[val]; !isAlreadyCalculated {
			remainingInputs = append(remainingInputs, val)
		}
	}

	// --------- STEP ----------
	// sending the availability queries
	//start = time.Now()
	nodeGatheringConclusion, err := n.sendBudgetedAvailabilityQueries(requestID, numberOfRequestedNodes)
	if err != nil {
		return nil, err
	}

	ticker := time.NewTicker(5 * time.Second)
	select {
	case <-n.terminationChannel:
		availableNodes := n.computationManager.getAvailableNodes(requestID)
		err = n.sendReservationCancellationMessages(requestID, availableNodes)
		return nil, err
	case <-nodeGatheringConclusion:

	case <-ticker.C:
		availableNodes := n.computationManager.getAvailableNodes(requestID)
		if uint(len(availableNodes)) < numberOfRequestedNodes {
			err = n.sendReservationCancellationMessages(requestID, availableNodes)
			return nil, xerrors.Errorf("Unable to reserve enough nodes. Desired nodes %v, obtained nodes %v",
				numberOfRequestedNodes, len(availableNodes))
		}
	}

	availableNodes := n.computationManager.getAvailableNodes(requestID)
	//elapsed = time.Since(start)
	//fmt.Println("Availability collection duration: ", elapsed.Seconds())

	// --------- STEP ----------
	// divide the work among the proposed nodes and send them computation orders
	var inputsPerNodeArray [][]string = make([][]string, len(availableNodes))
	i := 0
	for j := 0; j < len(remainingInputs); j++ {
		i = i % len(availableNodes)
		inputsPerNodeArray[i] = append(inputsPerNodeArray[i], remainingInputs[j])
		i++
	}

	inputsPerNode := make(map[string][]string)
	for i, address := range availableNodes {
		inputsPerNode[address] = inputsPerNodeArray[i]
	}
	err = n.sendComputationOrders(requestID, executionArgs, fileExtension, executable, inputsPerNode)
	if err != nil {
		return nil, err
	}
	// --------- STEP ----------
	// wait for all computations to conclude
	println("Waiting for computation to finish")
	//start = time.Now()
	remoteComputationsResults := <-computationConclusionChannel
	println("All computation is done")

	var sb strings.Builder
	for _, input := range inputData {
		sb.WriteString(input + " " + remoteComputationsResults[input] + "\n")
	}
	//elapsed = time.Since(start)
	//fmt.Println("Pure execution duration: ", elapsed.Seconds())

	// --------- STEP -------
	// calculate the cost of the total operation and how much to send to everyone
	//start = time.Now()
	costMap := make(map[string]float64)
	totalCostTruncated := float64(int(totalCost*1000)) / 1000

	fmt.Printf("Total cost: %v\n", totalCostTruncated)

	costMap[n.socket.GetAddress()] -= totalCostTruncated

	for nodeAddress, inputs := range inputsPerNode {
		costMap[nodeAddress] = costPerUnit*float64(len(inputs)) + 1 // there's a base cost of 1 per node
	}
	fmt.Println("Starting budget update...")
	err = n.UpdateBudget(requestID, costMap)
	if err != nil {
		return nil, err
	}

	//elapsed = time.Since(start)
	//fmt.Println("Paxos duration: ", elapsed.Seconds())

	return []byte(sb.String()), nil
}

// saves the executable code as a file
// returns filename
func saveExecutable(executable []byte, extension string) (string, error) {
	code := string(executable)
	timestamp := time.Now().Unix()
	filename := EXECUTABLE_LOCATION + strconv.FormatInt(timestamp, 10) + extension
	file, err := os.Create(filename)
	defer file.Close()

	_, err = file.WriteString(code)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func splitData(data []byte) []string {
	dataString := string(data)
	splitData := strings.Split(strings.ReplaceAll(dataString, "\r\n", "\n"), "\n")
	return splitData
}

// return cost of executing each data point that has not already been evaluated,
//results of already performed computations
func estimateCost(filename string, executionArgs []string, inputsArray []string) (float64, map[string]string, error) {
	rand.Seed(time.Now().UnixNano())
	results := make(map[string]string)
	separatedData := make([]string, len(inputsArray))
	copy(separatedData, inputsArray) // make a copy to avoid shuffling the source array
	rand.Shuffle(len(separatedData), func(i, j int) { separatedData[i], separatedData[j] = separatedData[j], separatedData[i] })
	firstArgs := separatedData[0]
	codeArgs := strings.Split(firstArgs, ",")
	app := executionArgs[0]
	args := make([]string, 0, 1+len(executionArgs)-1+len(codeArgs))
	args = append(args, executionArgs[1:]...)
	args = append(args, filename)
	args = append(args, codeArgs...)
	//fmt.Println(args)
	start := time.Now()

	output, err := exec.Command(app, args...).Output()
	if err != nil {
		return 0, nil, err
	}

	elapsed := time.Since(start)
	results[separatedData[0]] = string(output)
	// if time elapsed was more than 5 use just one data point to estimate cost
	if elapsed.Seconds() >= 5 {
		return math.RoundToEven(elapsed.Seconds()),
			results, nil
	}

	// if one execution took less than 5 seconds
	// calculate the average duration, by measuring two more data points
	durations := []time.Duration{elapsed}

	for i := 1; i < 3; i++ {
		firstArgs = separatedData[i]
		codeArgs := strings.Split(firstArgs, ",")
		app := executionArgs[0]
		args := make([]string, 0, 1+len(executionArgs)-1+len(codeArgs))
		args = append(args, executionArgs[1:]...)
		args = append(args, filename)
		args = append(args, codeArgs...)
		//fmt.Println(args)
		start = time.Now()

		output, err = exec.Command(app, args...).Output()
		if err != nil {
			return 0, nil, err
		}
		elapsed = time.Since(start)
		results[separatedData[i]] = string(output)
		durations = append(durations, elapsed)
	}

	// calculate the average duration
	sum := 0.0
	for _, duration := range durations {
		sum += duration.Seconds()
	}
	averageDuration := sum / 3.0

	// if the duration was more than 1 second, just round cost to integer
	if averageDuration >= 1 {
		//println("Average duration of running one sample above 1 ", averageDuration)
		return math.RoundToEven(averageDuration), results, nil
	} else {
		return math.Round(averageDuration*1000) / 1000, results, nil
	}
}

// divides the budget evenly among the multiple neighbours and sens an availability query to them
func (n *node) sendBudgetedAvailabilityQueries(requestID string, budget uint) (chan struct{}, error) {
	neighbourList := n.routingTable.neighbourList
	if len(neighbourList) == 0 {
		return nil, xerrors.New("The node has no neighbours")
	}

	availabilityChannel := n.computationManager.createAvailabilityChannel(requestID, budget)

	if int(budget) <= len(neighbourList) {
		for _, neighbourAddress := range neighbourList {
			pkt, err := n.createAvailabilityQueryPacket(requestID, 1, neighbourAddress, n.socket.GetAddress())
			if err != nil {
				return nil, err
			}
			err = n.socket.Send(neighbourAddress, pkt, 0)
			if err != nil {
				return nil, err
			}

			budget--
			if budget <= 0 {
				break
			}
		}
	} else {
		budgets := make([]uint, len(neighbourList))
		i := 0
		for budget > 0 {
			i = i % len(neighbourList)
			budgets[i]++
			budget--
			i++
		}

		for i, neighbourAddress := range neighbourList {
			pkt, err := n.createAvailabilityQueryPacket(requestID, budgets[i], neighbourAddress, n.socket.GetAddress())
			if err != nil {
				return nil, err
			}
			err = n.socket.Send(neighbourAddress, pkt, 0)
			if err != nil {
				return nil, err
			}
		}
	}

	return availabilityChannel, nil
}

// tells every reserved node to free up their resources
func (n *node) sendReservationCancellationMessages(requestID string, addressList []string) error {
	for _, destinationAddress := range addressList {
		reservationCancellation := types.ReservationCancellationMessage{
			RequestID: requestID,
		}

		reservationCancellationMarshaled, err := n.registry.MarshalMessage(reservationCancellation)
		if err != nil {
			return err
		}

		reservationCancellationHeader := transport.NewHeader(n.socket.GetAddress(), n.socket.GetAddress(), destinationAddress, 0)
		pkt := transport.Packet{
			Header: &reservationCancellationHeader,
			Msg:    &reservationCancellationMarshaled,
		}
		fmt.Println("sent cancellation to ", destinationAddress)
		err = n.socket.Send(destinationAddress, pkt, 0)
		if err != nil {
			return err
		}
	}

	return nil
}

// send a computation order to each of the promised nodes with the inputs they should process
func (n *node) sendComputationOrders(requestID string, executionArgs []string, fileExtension string,
	executable []byte, inputsPerNode map[string][]string) error {
	for nodeAddress, inputs := range inputsPerNode {
		computationOrder := types.ComputationOrderMessage{
			RequestID:     requestID,
			Executable:    executable,
			FileExtension: fileExtension,
			ExecutionArgs: executionArgs,
			Inputs:        inputs,
		}

		computationOrderMarshaled, err := n.registry.MarshalMessage(computationOrder)
		if err != nil {
			return err
		}
		computationOrderHeader := transport.NewHeader(n.socket.GetAddress(), n.socket.GetAddress(), nodeAddress, 0)
		pkt := transport.Packet{
			Header: &computationOrderHeader,
			Msg:    &computationOrderMarshaled,
		}
		err = n.socket.Send(nodeAddress, pkt, 0)
		if err != nil {
			return err
		}
	}
	return nil
}

func (n *node) createAvailabilityQueryPacket(requestID string, budget uint, peerAddress string, origin string) (transport.Packet, error) {
	availabilityRequest := types.AvailabilityQueryMessage{
		RequestID:      requestID,
		Source:         origin,
		Budget:         budget,
		AlreadyVisited: map[string]bool{origin: true},
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
