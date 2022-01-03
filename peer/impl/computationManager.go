package impl

import (
	"sync"
)

type responseMap map[string][]string // request ID, IPs of the nodes who said they're available
type computationManager struct {
	balance            int
	requestBeingServed string // id of the computation that reserved the node
	// this is used to stop propagating queries of the same id which have lower budgets, since they'll never reach as far
	// as the one previously sent with an higher budget
	maxBudgetForwarded map[string]uint // key: requestID, value: max budget forwarded
	responseMap                        // used for requests created by this node
	mutex              sync.Mutex
}

func newComputationManager() *computationManager {
	return &computationManager{
		balance:            50,
		maxBudgetForwarded: make(map[string]uint),
		responseMap:        make(responseMap),
		requestBeingServed: "",
		mutex:              sync.Mutex{},
	}
}

func (cm *computationManager) registerResponse(requestID string, nodeAddress string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.responseMap[requestID] = append(cm.responseMap[requestID], nodeAddress)
}

func (cm *computationManager) getResponseList(requestID string) []string {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	return cm.responseMap[requestID]
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

	// try to reserve if it wasn't reserved
	if cm.requestBeingServed == "" {
		cm.requestBeingServed = requestID
		cm.maxBudgetForwarded[requestID] = requestBudget - 1
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

	return cm.responseMap[requestID]
}
