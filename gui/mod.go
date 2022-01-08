package main

import (
	"flag"
	"fmt"
	"github.com/rs/xid"
	"go.dedis.ch/cs438/peer"
	"go.dedis.ch/cs438/peer/impl"
	"go.dedis.ch/cs438/registry/standard"
	"go.dedis.ch/cs438/storage/inmemory"
	"go.dedis.ch/cs438/transport/udp"
	"golang.org/x/xerrors"
	"time"
)

var totalNumberOfPeers = uint(2)
var count = uint(0)
var centralNodeAddress = "127.0.0.1:12345"

func main() {

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

		for {

		}
		return
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

		AntiEntropyInterval: time.Second,
		HeartbeatInterval:   time.Second,
		AckTimeout:          3 * time.Second,
		ContinueMongering:   0.5,

		ChunkSize: 8192,
		BackoffDataRequest: peer.Backoff{
			//2s 2 5
			Initial: 2 * time.Second,
			Factor:  2,
			Retry:   5,
		},
		Storage: storage,

		TotalPeers: totalNumberOfPeers,
		PaxosThreshold: func(u uint) int {
			return int(u/2 + 1)
		},
		PaxosID:            count,
		PaxosProposerRetry: 7 * time.Second,
		InitialBudget:      10,
	}

	return peerFactory(conf)
}
