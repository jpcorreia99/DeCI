package impl

import (
	"fmt"
)

func (n *node) UpdateBudget(computationID string, participantAmountMap map[string]int) error {

	s := encodeMapToString(participantAmountMap)

	return n.Tag(computationID, s)
}

func encodeMapToString(participantAmountMap map[string]int) string {
	s := "{"
	for key, val := range participantAmountMap {
		temp := fmt.Sprintf("\"%s\" : %v,", key, val)
		s += temp
	}
	s = s[:len(s)-1] + "}"

	return s
}
