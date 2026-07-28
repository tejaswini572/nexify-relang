package main

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"time"
)

func token() string {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		sum := sha1.Sum([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
		return hex.EncodeToString(sum[:])
	}
	return hex.EncodeToString(b)
}
