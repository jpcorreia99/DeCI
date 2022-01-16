package peer

import "sync"

type ConcurrentUint struct {
	amount uint // packet ID, channel to signal it has been acked back
	mutex  sync.Mutex
}

func NewConcurrentUint(init uint) *ConcurrentUint {
	return &ConcurrentUint{
		amount: init, // packetId, channel to signal its receipt
		mutex:  sync.Mutex{},
	}
}

func (cb *ConcurrentUint) Get() uint {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	res := cb.amount

	return res
}

func (cb *ConcurrentUint) Set(flag uint) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.amount = flag
}

func (cb *ConcurrentUint) Increase() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.amount += 1
}
