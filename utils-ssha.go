package main

import (
	"github.com/jsimonetti/pwscheme/ssha512"
)

// Encode encodes the []byte of raw password
func SSHAEncode(rawPassPhrase string) (string, error) {
	return ssha512.Generate(rawPassPhrase, 16)
}
