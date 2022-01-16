package impl

import "sync"

// used by the broadcasting function to wait until the ack has been received,
type ConcurrentPeerMap struct {
	ackMap map[string]bool // packet ID, channel to signal it has been acked back
	mutex  sync.Mutex
}

func newConcurrentPeerMap() *ConcurrentPeerMap {
	return &ConcurrentPeerMap{
		ackMap: make(map[string]bool), // packetId, channel to signal its receipt
		mutex:  sync.Mutex{},
	}
}

func (cpm *ConcurrentPeerMap) add(peerAddress string) {
	cpm.mutex.Lock()
	defer cpm.mutex.Unlock()

	cpm.ackMap[peerAddress] = true
}

func (cpm *ConcurrentPeerMap) contains(peerAddress string) bool {
	cpm.mutex.Lock()
	defer cpm.mutex.Unlock()

	_, ok := cpm.ackMap[peerAddress]

	return ok
}
