package impl

import "sync"

type ConcurrentJoinAckMap struct {
	ackMap map[string]bool // packet ID, channel to signal it has been acked back
	mutex  sync.Mutex
}

func newConcurrentJoinAckMap() *ConcurrentJoinAckMap {
	return &ConcurrentJoinAckMap{
		ackMap: make(map[string]bool), // packetId, channel to signal its receipt
		mutex:  sync.Mutex{},
	}
}

func (cjam *ConcurrentJoinAckMap) add(peerAddress string) {
	cjam.mutex.Lock()
	defer cjam.mutex.Unlock()

	cjam.ackMap[peerAddress] = true
}

func (cjam *ConcurrentJoinAckMap) contains(peerAddress string) bool {
	cjam.mutex.Lock()
	defer cjam.mutex.Unlock()

	_, ok := cjam.ackMap[peerAddress]

	return ok
}

func (cjam *ConcurrentJoinAckMap) getSize() int {
	cjam.mutex.Lock()
	defer cjam.mutex.Unlock()

	res := len(cjam.ackMap)

	return res
}
