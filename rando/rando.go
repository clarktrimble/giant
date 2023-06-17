// Package rando enshrines method 8 from https://stackoverflow.com/questions/22892120
package rando

import (
	"math/rand"
	"time"
	"unsafe"
)

// Todo: break down and push into tiny mod
// Todo: provide for passing in a seed for unit
// Todo: provide for passing in char set

// Rando generates a random string of length n
func Rando(n int) string {

	return randStringBytesMaskImprSrcUnsafe(n)
}

// unexported

// want digits as well por favor! 2^6 is 64 so appending digits should be good
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func randStringBytesMaskImprSrcUnsafe(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}
