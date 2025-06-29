package utils

import (
	"fmt"
	"math/rand"
)

func MakeRandomHexIdentifier(minLen, maxLen int) string {
	length := minLen + rand.Intn(maxLen-minLen+1)
	randomInt := rand.Int63()
	hexStr := fmt.Sprintf("%X", randomInt)
	if len(hexStr) > length {
		hexStr = hexStr[:length]
	}
	return hexStr
}
