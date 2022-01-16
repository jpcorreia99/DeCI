package impl

import "sync"

type ConcurrentBool struct {
	flag  bool // packet ID, channel to signal it has been acked back
	mutex sync.Mutex
}

func newConcurrentBool() *ConcurrentBool {
	return &ConcurrentBool{
		flag:  false, // packetId, channel to signal its receipt
		mutex: sync.Mutex{},
	}
}

func (cb *ConcurrentBool) get() bool {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	res := cb.flag

	return res
}

func (cb *ConcurrentBool) set(flag bool) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.flag = flag
}
