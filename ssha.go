package main

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// Encode encodes the []byte of raw password
func SSHAEncode(rawPassPhrase []byte) string {
	hash := makeSSHAHash(rawPassPhrase, makeSalt())
	b64 := base64.StdEncoding.EncodeToString(hash)
	return fmt.Sprintf("{ssha}%s", b64)
}

// makeSalt make a 32 byte array containing random bytes.
func makeSalt() []byte {
	sbytes := make([]byte, 32)
	_, err := rand.Read(sbytes)
	if err != nil {
		log.Panicf("Could not read random bytes: %s", err)
	}
	return sbytes
}

// makeSSHAHash make hasing using SHA-1 with salt. This is not the final output though. You need to append {SSHA} string with base64 of this hash.
func makeSSHAHash(passphrase, salt []byte) []byte {
	sha := sha1.New()
	sha.Write(passphrase)
	sha.Write(salt)

	h := sha.Sum(nil)
	return append(h, salt...)
}
