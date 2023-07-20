package main

import (
	"fmt"
	"log"
	// "bytes"
	// "crypto/rand"
	// "encoding/binary"
	// "encoding/hex"
	// "fmt"
	// "html/template"
	// "log"
	// "net/http"
	// "regexp"
	// "strings"
	// "github.com/emersion/go-sasl"
	// "github.com/emersion/go-smtp"
	// "github.com/gorilla/mux"
	// "golang.org/x/crypto/argon2"
)

type NewUser struct {
	DN          string
	CN          string
	GivenName   string
	DisplayName string
	Mail        string
	SN          string
	UID         string
}

func addNewUser(newUser NewUser) {
	log.Printf(fmt.Sprint("Adding New User"))
}
