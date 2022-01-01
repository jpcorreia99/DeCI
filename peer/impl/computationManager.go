package impl

import (
	"go.dedis.ch/cs438/types"
	"sync"
)

type responseMap map[string][]types.AvailabilityResponseMessage // request ID, responses of nodes sayng they're available
type computationManager struct {
	balance            int
	requestBeingServed string // id of the computation that reserved the node
	responseMap               // used for requests created by this node
	mutex              sync.Mutex
}

func newComputationManager() *computationManager {
	return &computationManager{
		balance:            50,
		responseMap:        make(responseMap),
		requestBeingServed: "",
		mutex:              sync.Mutex{},
	}
}

func (cm *computationManager) registerResponse(responseMsg types.AvailabilityResponseMessage) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.responseMap[responseMsg.RequestID] = append(cm.responseMap[responseMsg.RequestID], responseMsg)
}

func (cm *computationManager) getResponseList(requestID string) []types.AvailabilityResponseMessage {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	return cm.responseMap[requestID]
}

// tries to reserve the node to execute the requested computation
// returns if it was alble to reserve and the current balance
func (cm *computationManager) tryReserve(requestID string) (bool, int) {
	//todo: launch a routine to unlock the computer if no response is received, maybe(?)
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if cm.requestBeingServed == "" {
		cm.requestBeingServed = requestID
		return true, cm.balance
	}
	return false, 0
}
