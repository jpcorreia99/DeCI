package impl

import (
	"fmt"
	"github.com/rs/xid"
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/types"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// todo: when we send a computation, also send the info about payment,
// if the payment was too far from reality, cost more or less
// todo: sort inputs to prevent sketchy order
// input the cost of the computation, the computation's id and the list of nodes that participated in the computation

func (n *node) Compute(executable []byte, data []byte) ([]byte, error) {

	// --------- STEP ----------
	// sending the availability queries
	budget := uint(6)
	requestID := xid.New().String()
	err := n.sendBudgetedAvailabilityQueries(requestID, budget)
	if err != nil {
		return nil, err
	}

	time.Sleep(5 * time.Second)
	fmt.Println("xd")
	return nil, nil

	filename, err := saveExecutable(executable)
	if err != nil {
		return nil, err
	}

	inputData := splitData(data)
	costPerUnit, alreadyCalculatedResults, err := estimateCost(filename, inputData)
	if err != nil {
		return nil, err
	}
	fmt.Println("not error")
	results := make(map[string]string)
	for i := 0; i < len(alreadyCalculatedResults); i++ {
		results[inputData[i]] = alreadyCalculatedResults[i]
	}

	// remove form the data the ones already calculated
	remainingData := inputData[len(alreadyCalculatedResults):]
	println(costPerUnit)
	fmt.Println(results) // todo: não está a dar print de tudo
	println(len(remainingData))
	return nil, nil

	/*for _, line := range remainingData {
		codeArgs := strings.Split(line, ",")
		app := "python"
		args := make([]string, 0, 1+len(codeArgs))
		args = append(args, filename)
		args = append(args, codeArgs...)
		fmt.Println(args)
		output, err := exec.Command(app, args...).Output()
		if err != nil {
			fmt.Println(":(")
			return nil, err
		}
		fmt.Print("output ", string(output))
	}

	return nil, nil*/
}

// saves the executable code as a go file
// returns filename
func saveExecutable(executable []byte) (string, error) {
	code := string(executable)
	timestamp := time.Now().Unix()
	filename := "executables/" + strconv.FormatInt(timestamp, 10)
	file, err := os.Create(filename)
	defer file.Close()
	// len variable captures the length
	// of the string written to the file.
	_, err = file.WriteString(code)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func splitData(data []byte) []string {
	dataString := string(data)
	splitData := strings.Split(strings.ReplaceAll(dataString, "\r\n", "\n"), "\n")
	rand.Shuffle(len(splitData), func(i, j int) { splitData[i], splitData[j] = splitData[j], splitData[i] })
	return splitData
}

// return cost of executing each data point that has not already been evaluated,
//results of already performed computations
func estimateCost(filename string, separatedData []string) (float64, []string, error) {
	firstArgs := separatedData[0]
	codeArgs := strings.Split(firstArgs, ",")
	app := "python"
	args := make([]string, 0, 1+len(codeArgs))
	args = append(args, filename)
	args = append(args, codeArgs...)
	fmt.Println(args)
	start := time.Now()

	output, err := exec.Command(app, args...).Output()
	if err != nil {
		return 0, nil, err
	}

	elapsed := time.Since(start)
	// if time elapsed was more than 5 use just one data point to estimate cost
	if elapsed.Seconds() >= 5 {
		println("average duration above 5: ", elapsed.Seconds())
		return math.RoundToEven(elapsed.Seconds()),
			[]string{string(output)}, nil
	}

	// if one execution took less than 5 seconds
	// calculate the average duration, by measuring two more data points
	results := []string{string(output)}
	durations := []time.Duration{elapsed}

	for i := 1; i < 3; i++ {
		firstArgs = separatedData[i]
		codeArgs := strings.Split(firstArgs, ",")
		app := "python"
		args := make([]string, 0, 1+len(codeArgs))
		args = append(args, filename)
		args = append(args, codeArgs...)
		fmt.Println(args)
		start = time.Now()

		output, err = exec.Command(app, args...).Output()
		if err != nil {
			return 0, nil, err
		}
		elapsed = time.Since(start)
		results = append(results, string(output))
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
		println("Average duration above 1 ", averageDuration)
		return math.RoundToEven(averageDuration), results, nil
	} else {
		println("Average duration under 1 ", averageDuration)
		return math.Round(averageDuration*1000) / 1000, results, nil
	}
}

// divides the budget evenly among the multiple neighbours and sens an availability query to them
func (n *node) sendBudgetedAvailabilityQueries(requestID string, budget uint) error {
	neighbourList := n.routingTable.neighbourList
	if len(neighbourList) == 0 {
		return nil
	}

	if int(budget) <= len(neighbourList) {
		for _, neighbourAddress := range neighbourList {
			pkt, err := n.createAvailabilityQueryPacket(requestID, 1, neighbourAddress, n.socket.GetAddress())
			if err != nil {
				return err
			}

			err = n.socket.Send(neighbourAddress, pkt, 0)
			if err != nil {
				return err
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

func (n *node) createAvailabilityQueryPacket(requestID string, budget uint, peerAddress string, origin string) (transport.Packet, error) {
	availabilityRequest := types.AvailabilityQueryMessage{
		RequestID: requestID,
		Source:    origin,
		Budget:    budget,
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
