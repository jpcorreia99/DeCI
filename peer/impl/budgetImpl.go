package impl

import (
	"fmt"
)

func (n *node) UpdateBudget(computationID string, participantAmountMap map[string]float64) error {

	s := encodeMapToString(participantAmountMap)

	return n.Tag(computationID, s)
}

func encodeMapToString(participantAmountMap map[string]float64) string {
	s := "{"
	for key, val := range participantAmountMap {
		temp := fmt.Sprintf("\"%s\" : %.2f,", key, val)
		s += temp
	}
	s = s[:len(s)-1] + "}"

	return s
}
