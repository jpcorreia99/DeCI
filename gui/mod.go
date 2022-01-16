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
	"os"
	"time"
)

func main() {
	isCentral := flag.Bool("central", false, "Create a central node.")
	address := flag.String("address", "127.0.0.1:0", "Create a node with a specific address.")
	data := flag.String("data", "", "Path to data file.")
	code := flag.String("code", "", "Path to code file.")
	resultPath := flag.String("result", "", "Path to result file.")
	nodes := flag.Uint("nodes", 0, "Number of requested computation nodes.")

	flag.Parse()

	if *isCentral {
		centralNode := createPeer("127.0.0.1:12345")
		fmt.Printf("Created central node at address: %s\n", centralNode.GetAddress())
		err := centralNode.Start()
		defer centralNode.Stop()
		if err != nil {
			fmt.Println(err)
			return
		}

		time.Sleep(time.Hour)
	}

	myNode := createPeer(*address)

	err := myNode.Start()
	defer myNode.Stop()
	if err != nil {
		println(err)
		return
	}

	fmt.Printf("Created a node at address %v\n", myNode.GetAddress())

	myNode.AddPeer("127.0.0.1:12345")

	err = myNode.JoinNetwork("127.0.0.1:12345")

	println("Joined the network!")

	if err != nil {
		println(err)
		return
	}

	if *data != "" && *code != "" && *nodes != 0 {

		dataRead, err := os.ReadFile(*data)
		if err != nil {
			println(err)
			return
		}

		codeRead, err := os.ReadFile(*code)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("Current budget: %v\n", myNode.GetBudget())
		println("Sending computation!")
		res, err2 := myNode.Compute(codeRead, []string{"python"}, ".py", dataRead, *nodes)
		fmt.Printf("New budget: %v\n", myNode.GetBudget())

		if err2 != nil {
			println(err2)
			return
		}

		resString := string(res)

		fmt.Printf("Result:\n%v\n", resString)

		if *resultPath != "" {
			file, err := os.Create(*resultPath)

			if err != nil {
				fmt.Println(err)
			} else {
				file.WriteString(resString)
			}
			err = file.Close()
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		println("Done!")

	} else {
		println("Sleeping and making money")
		time.Sleep(time.Second * 600)
	}

}

func createPeer(address string) peer.Peer {
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

		TotalPeers: peer.NewConcurrentUint(uint(1)),
		PaxosThreshold: func(u uint) int {
			return int(u/2 + 1)
		},
		PaxosID:            uint(xid.New().Counter()),
		PaxosProposerRetry: 7 * time.Second,
		InitialBudget:      100,
	}

	return peerFactory(conf)
}
