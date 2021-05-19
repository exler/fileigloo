package random

import (
	"math/rand"
	"time"
)

const characters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func String(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = characters[rand.Int63()%int64(len(characters))]
	}
	return string(b)
}
