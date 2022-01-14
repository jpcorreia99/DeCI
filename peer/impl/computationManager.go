package impl

import (
	"sync"
	"time"
)

type computationManager struct {
	balance               int
	requestBeingServed    string // id of the computation that reserved the node -> not used by computations issued by the node, only for the ones received from external sources
	hasReceivedExecutable bool   // boolean indicating if, after being reserved, the node has already received the executable code

	// maxBudgetForwarded - this is used to stop propagating queries of the same id which have lower budgets, since they'll never reach as far
	// as the one previously sent with an higher budget
	maxBudgetForwarded                map[string]uint                   // key: requestID, value: max budget forwarded
	cancelledComputations             map[string]bool                   // set of IDs of cancelled computations, to avoid reserving the node again
	availabilityMap                   map[string][]string               // key: requesstID, IPs of the nodes who said they're available. used for requests created by this node
	pretendedNodesCountMap            map[string]uint                   // key; requestID, value: number of nodes desired for that computation
	availabilityCollectionChannelMap  map[string]chan struct{}          // key: requestID, value: channel which is triggered when the pretended number of available nodes is received
	computationFinalizationChannelMap map[string]chan map[string]string // key: request ID, value: channel that send the final result of all the remote computations
	remainingInputMap                 map[string]uint                   // key: request ID, value: number of inputs whose result has not yet been received
	resultsMap                        map[string]map[string]string      // key: requestID, value: map with inputs and their results. Used for requests created by this node
	mutex                             sync.Mutex
}

func newComputationManager() *computationManager {
	return &computationManager{
		balance:                           50,
		maxBudgetForwarded:                make(map[string]uint),
		pretendedNodesCountMap:            make(map[string]uint),
		availabilityMap:                   make(map[string][]string),
		availabilityCollectionChannelMap:  make(map[string]chan struct{}),
		computationFinalizationChannelMap: make(map[string]chan map[string]string),
		remainingInputMap:                 make(map[string]uint),
		resultsMap:                        make(map[string]map[string]string),
		cancelledComputations:             make(map[string]bool),
		requestBeingServed:                "",
		hasReceivedExecutable:             false,
		mutex:                             sync.Mutex{},
	}
}

func (cm *computationManager) registerAvailability(requestID string, nodeAddress string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.availabilityMap[requestID] = append(cm.availabilityMap[requestID], nodeAddress)

	if uint(len(cm.availabilityMap[requestID])) == cm.pretendedNodesCountMap[requestID] {
		close(cm.availabilityCollectionChannelMap[requestID])
	}
}

func (cm *computationManager) getResponseList(requestID string) []string {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	return cm.availabilityMap[requestID]
}

// tries to reserve the node to execute the requested computation
// 1st bool: was able to reserve
// 2nd bool: if the request should be propagated
/* order of comparisons
1-  if no computation was being executed, it can be reserved
	1.1 if after reducing the budget by one it is still above 1, it should be forwarded, else don't forward
2 - if the request if for the same requestID the node has already accepted:
	2.1 if the budget of this request is bigger than the one last forwarded, it can't be reserved,
		but the request should be forwarded
	2.2 if it is lower, don't reserve not forward
3 - if this request does not have the same ID as the one being served:
	3.1 forward it if it was never seen or if the previous forwarded budget was smaller than the current one
	3.2 otherwise don't forward it
*/
func (cm *computationManager) tryReserve(requestID string, requestBudget uint) (bool, bool) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if _, requestWasCancelledPreviously := cm.cancelledComputations[requestID]; requestWasCancelledPreviously {
		return false, false
	}

	// try to reserve if it wasn't reserved
	if cm.requestBeingServed == "" {
		cm.requestBeingServed = requestID
		cm.maxBudgetForwarded[requestID] = requestBudget - 1
		go cm.selfDeReservation(requestID) // launches a routine that removes the reservation if after a given amount of time no executable has been received
		if requestBudget-1 > 0 {
			return true, true
		} else {
			return true, false
		}
	}

	if cm.requestBeingServed == requestID {
		if requestBudget > cm.maxBudgetForwarded[requestID] {
			cm.maxBudgetForwarded[requestID] = requestBudget
			return false, true
		} else {
			return false, false
		}
	} else {
		if maxBudgetServed, requestIsRegistered := cm.maxBudgetForwarded[requestID]; requestIsRegistered && maxBudgetServed < requestBudget {
			return false, true
		} else if requestIsRegistered && maxBudgetServed >= requestBudget {
			return false, false
		} else {
			cm.maxBudgetForwarded[requestID] = requestBudget
			return false, true
		}
	}
}

// returns the IPs of the nodes who promised to execute this computation
func (cm *computationManager) getAvailableNodes(requestID string) []string {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	return cm.availabilityMap[requestID]
}

func (cm *computationManager) cancelReservation(requestID string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if cm.requestBeingServed == requestID {
		cm.requestBeingServed = ""
		cm.hasReceivedExecutable = false
	}

	cm.cancelledComputations[requestID] = true
}

// checks if the id of a received computation order matches the one that this node has promised
// if it was, turn the boolean that signals a computation has been received to true
func (cm *computationManager) signalExecutableReceived(requestID string) bool {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	wasIdRequested := cm.requestBeingServed == requestID
	if !wasIdRequested {
		return false
	} else {
		cm.hasReceivedExecutable = true
		return true
	}
}

func (cm *computationManager) registerComputation(requestID string, numberOfInputs uint) chan map[string]string {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	newChannel := make(chan map[string]string)
	cm.computationFinalizationChannelMap[requestID] = newChannel
	cm.resultsMap[requestID] = make(map[string]string)
	cm.remainingInputMap[requestID] = numberOfInputs

	return newChannel
}

func (cm *computationManager) createAvailabilityChannel(requestID string, numberOfPretendedNodes uint) chan struct{} {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.pretendedNodesCountMap[requestID] = numberOfPretendedNodes
	availabilityChannel := make(chan struct{})
	cm.availabilityCollectionChannelMap[requestID] = availabilityChannel

	return availabilityChannel
}

// todo: if we want to be extra safe we can ensure the results we get are from the node which is supposed to have sent them
func (cm *computationManager) storeResults(requestID string, results map[string]string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	for input, output := range results {
		cm.resultsMap[requestID][input] = output
	}

	cm.remainingInputMap[requestID] -= uint(len(results))
	if cm.remainingInputMap[requestID] == 0 {
		cm.computationFinalizationChannelMap[requestID] <- cm.resultsMap[requestID]
		close(cm.computationFinalizationChannelMap[requestID])
		delete(cm.computationFinalizationChannelMap, requestID)
		delete(cm.resultsMap, requestID)
		delete(cm.availabilityMap, requestID)
		delete(cm.remainingInputMap, requestID)
		delete(cm.availabilityCollectionChannelMap, requestID)
		delete(cm.pretendedNodesCountMap, requestID)
	}
}

// gets called as a routine. After a given amount of time, if the node hasn't yet receuived the executable for the
// request is has been reserved for it deReserves itself
func (cm *computationManager) selfDeReservation(requestID string) {
	time.Sleep(10 * time.Second)

	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if cm.requestBeingServed == requestID && !cm.hasReceivedExecutable {
		cm.requestBeingServed = ""
	}
}
