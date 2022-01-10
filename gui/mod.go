package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/rs/xid"
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/peer/impl"
	"go.dedis.ch/cs438/registry/standard"
	"go.dedis.ch/cs438/storage/inmemory"
	"go.dedis.ch/cs438/transport"
	"go.dedis.ch/cs438/transport/udp"
	"go.dedis.ch/cs438/types"
	"golang.org/x/xerrors"
	"os"
	"strconv"
	"time"
)

var count = uint(0)
var centralNodeAddress = "127.0.0.1:12345"
var totalNumberOfPeers = 11

func main2() {

	isCentral := flag.Bool("central", false, "Create a central node.")
	//address := flag.String("address", "127.0.0.1:0", "Create a node with a specific address.")
	flag.Parse()

	if *isCentral {
		fmt.Printf("Creating central node at address: %s\n", centralNodeAddress)
		centralNode := createPeer(centralNodeAddress)
		err := centralNode.Start()
		if err != nil {
			fmt.Println(err)
			return
		}

		select {}
		//err = centralNode.Stop()
		//if err != nil {
		//	fmt.Println(err)
		//	return
		//}
	}

	node1 := createPeer("127.0.0.1:0")
	node2 := createPeer("127.0.0.1:0")

	//node1.AddPeer()

	err := node1.Start()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = node2.Start()
	if err != nil {
		fmt.Println(err)
		return
	}

	computationID := xid.New().String()

	fmt.Printf("Issuing new computation with ID %v\n", computationID)

	participantAmountMap := make(map[string]int)
	participantAmountMap["42"] = 1
	participantAmountMap["69"] = 2

	s := encodeMapToString(participantAmountMap)

	err = node1.Tag(computationID, s)

	err = node1.Stop()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("closed")

	//s := encodeMapToString(participantAmountMap)
	//
	//
	//decodeMap := make(map[string]int)
	//
	//err := json.Unmarshal([]byte(s), &decodeMap)
	//
	//if err != nil {
	//	fmt.Printf("Error %v\n", err)
	//}
	//println("All good")
	//my := 17
	//for key, value := range decodeMap {
	//	if key == "42"{
	//		my -= value
	//	}
	//}
	//
	//fmt.Printf("My : %v\n", my)
}

func main() {
	var nodeList []peer.Peer
	for i := 0; i < totalNumberOfPeers; i++ {
		address := 11110
		stringAddress := "127.0.0.1:" + strconv.Itoa(address+i)
		println(stringAddress)
		peerNode := createPeer(stringAddress)
		defer peerNode.Stop()
		nodeList = append(nodeList, peerNode)
	}

	for i := 0; i < totalNumberOfPeers; i++ {
		for j := 0; j < totalNumberOfPeers; j++ {
			if i != j {
				address := 11110
				stringAddress := "127.0.0.1:" + strconv.Itoa(address+j)
				nodeList[i].AddPeer(stringAddress)
			}
		}
	}

	for i := 0; i < totalNumberOfPeers; i++ {
		err := nodeList[i].Start()
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	code, err := os.ReadFile("executables/dupper.py")
	if err != nil {
		fmt.Println(err)
		return
	}

	data, err := os.ReadFile("executables/numbers.txt")

	start := time.Now()
	_, err = nodeList[0].Compute(code, data, 10)

	elapsed := time.Since(start)
	fmt.Println("duration: ", elapsed.Seconds())
	if err != nil {
		fmt.Println(err)
		return
	}

	/*for i := 0; i < totalNumberOfPeers; i++ {
		fmt.Println("closing ", i)
		err := nodeList[i].Stop()
		if err != nil {
			fmt.Println(err)
			return
		}
	}*/
}
func main3() {
	node1 := createPeer("127.0.0.1:12345")
	node2 := createPeer("127.0.0.1:54321")
	node3 := createPeer("127.0.0.1:55555")
	node4 := createPeer("127.0.0.1:11111")

	/*
					    4    -  6
		             /     \  /
		    		1		3
		             \	   /
		   \			2
	*/
	node1.AddPeer("127.0.0.1:54321") // 1 -> 2
	node2.AddPeer("127.0.0.1:12345") // 2 -> 1
	node2.AddPeer("127.0.0.1:55555") // 2 -> 3
	node3.AddPeer("127.0.0.1:54321") // 3 -> 2
	node1.AddPeer("127.0.0.1:1111")  // 1 -> 4
	node3.AddPeer("127.0.0.1:1111")  // 3 -> 4
	node4.AddPeer("127.0.0.1:12345") // 4 -> 1
	node4.AddPeer("127.0.0.1:55555") // 4 -> 3

	err := node1.Start()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = node2.Start()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = node3.Start()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = node4.Start()
	if err != nil {
		fmt.Println(err)
		return
	}

	chat := types.ChatMessage{
		Message: "Hello from node 1",
	}

	buf, err := json.Marshal(&chat)
	if err != nil {
		fmt.Println(err)
	}

	chatMsg := transport.Message{
		Type:    chat.Name(),
		Payload: buf,
	}
	err = node1.Broadcast(chatMsg)
	if err != nil {
		fmt.Println(err)
		return
	}

	code, err := os.ReadFile("executables/cracker.py")
	if err != nil {
		fmt.Println(err)
		return
	}

	data, err := os.ReadFile("executables/data.txt")

	_, err = node2.Compute(code, data, 3)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = node1.Stop()
	fmt.Println("a")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("b")
	err = node2.Stop()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("c")
	err = node3.Stop()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("d")
	err = node4.Stop()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("closed")
}

func encodeMapToString(participantAmountMap map[string]int) string {
	s := "{"
	for key, val := range participantAmountMap {
		// Convert each key/value pair in m to a string
		temp := fmt.Sprintf("\"%s\" : %v,", key, val)
		// Do whatever you want to do with the string;
		// in this example I just print out each of them.
		s += temp
	}
	s = s[:len(s)-1] + "}"

	fmt.Printf("The encoded string %v\n", s)

	return s
}

//func decodeMapFromString(s string) (map[string]int, error){
//	decodedMap := make(map[string]int)
//
//	err := json.Unmarshal([]byte(s), &decodedMap)
//
//	if err != nil {
//		return nil, err
//	}
//
//	return decodedMap, nil
//}

func createPeer(address string) peer.Peer {
	count++
	var peerFactory = impl.NewPeer

	trans := udp.NewUDP()

	sock, err := trans.CreateSocket(address)
	if err != nil {
		fmt.Println(xerrors.Errorf("failed to create socket"))
		return nil
	}

	var storage = inmemory.NewPersistency()

	conf := peer.Configuration{
		Socket:          sock,
		MessageRegistry: standard.NewRegistry(),

		//AntiEntropyInterval: 6 * time.Second,
		AntiEntropyInterval: 0,
		//HeartbeatInterval: 5 * time.Second,
		HeartbeatInterval: 0,
		AckTimeout:        3 * time.Second,
		ContinueMongering: 0.5,

		ChunkSize: 8192,
		BackoffDataRequest: peer.Backoff{
			//2s 2 5
			Initial: 2 * time.Second,
			Factor:  2,
			Retry:   5,
		},
		Storage: storage,

		TotalPeers: uint(totalNumberOfPeers),
		PaxosThreshold: func(u uint) int {
			return int(u/2 + 1)
		},
		PaxosID:            count,
		PaxosProposerRetry: 7 * time.Second,
		InitialBudget:      100,
	}

	return peerFactory(conf)
}
