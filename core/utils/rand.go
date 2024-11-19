package utils

import (
	"math/rand"
	"sync"
	"time"
)

var (
	randGenerate *rand.Rand
	randMu       sync.Mutex
)

func init() {
	seed := time.Now().UnixNano()
	source := rand.NewSource(seed)
	randGenerate = rand.New(source)
}

const (
	letters    = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	intLetters = "0123456789"
)

func randBase(n int, letters string) string {
	lettersLen := len(letters)
	if lettersLen == 0 {
		return ""
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[RandIntN(lettersLen)]
	}
	return string(b)
}

func RandStr(n int) string {
	return randBase(n, letters)
}

func RandIntStr(n int) string {
	return randBase(n, intLetters)
}

func RandIntN(n int) int {
	randMu.Lock()
	defer randMu.Unlock()
	return randGenerate.Intn(n)
}

func RandFloat64() float64 {
	randMu.Lock()
	defer randMu.Unlock()
	return randGenerate.Float64()
}
