package main

import (
  "crypto/rand"
  "crypto/sha256"
  "fmt"
)

func newCheckToken() (string, error) {
	c := 10
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

  return fmt.Sprintf("%s", sha256.Sum256(b)), nil
}
