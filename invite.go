package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/go-ldap/ldap/v3"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/argon2"
)

var EMAIL_REGEXP = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func checkInviterLogin(w http.ResponseWriter, r *http.Request) *LoginStatus {
	login := checkLogin(w, r)
	if login == nil {
		return nil
	}

	if !login.CanInvite {
		http.Error(w, "Not authorized to invite new users.", http.StatusUnauthorized)
		return nil
	}

	return login
}

// New account creation directly from interface

func handleInviteNewAccount(w http.ResponseWriter, r *http.Request) {
	login := checkInviterLogin(w, r)
	if login == nil {
		return
	}

	handleNewAccount(w, r, login.conn, login.Info.DN)
}

// New account creation using code

func handleInvitationCode(w http.ResponseWriter, r *http.Request) {
	code := mux.Vars(r)["code"]
	code_id, code_pw := readCode(code)

	l := ldapOpen(w)
	if l == nil {
		return
	}

	inviteDn := config.InvitationNameAttr + "=" + code_id + "," + config.InvitationBaseDN
	err := l.Bind(inviteDn, code_pw)
	if err != nil {
		templateInviteInvalidCode := getTemplate("invite_invalid_code.html")
		templateInviteInvalidCode.Execute(w, nil)
		return
	}

	sReq := ldap.NewSearchRequest(
		inviteDn,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(objectclass=*)"),
		[]string{"dn", "creatorsname"},
		nil)
	sr, err := l.Search(sReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(sr.Entries) != 1 {
		http.Error(w, fmt.Sprintf("Expected 1 entry, got %d", len(sr.Entries)), http.StatusInternalServerError)
		return
	}

	invitedBy := sr.Entries[0].GetAttributeValue("creatorsname")

	if handleNewAccount(w, r, l, invitedBy) {
		del_req := ldap.NewDelRequest(inviteDn, nil)
		err = l.Del(del_req)
		if err != nil {
			log.Printf("Could not delete invitation %s: %s", inviteDn, err)
		}
	}
}

// Common functions for new account

type NewAccountData struct {
	Username    string
	DisplayName string
	GivenName   string
	Surname     string

	ErrorUsernameTaken    bool
	ErrorInvalidUsername  bool
	ErrorPasswordTooShort bool
	ErrorPasswordMismatch bool
	ErrorMessage          string
	WarningMessage        string
	Success               bool
}

func handleNewAccount(w http.ResponseWriter, r *http.Request, l *ldap.Conn, invitedBy string) bool {
	templateInviteNewAccount := getTemplate("invite_new_account.html")

	data := &NewAccountData{}

	if r.Method == "POST" {
		r.ParseForm()

		data.Username = strings.TrimSpace(strings.Join(r.Form["username"], ""))
		data.DisplayName = strings.TrimSpace(strings.Join(r.Form["displayname"], ""))
		data.GivenName = strings.TrimSpace(strings.Join(r.Form["givenname"], ""))
		data.Surname = strings.TrimSpace(strings.Join(r.Form["surname"], ""))

		password1 := strings.Join(r.Form["password"], "")
		password2 := strings.Join(r.Form["password2"], "")

		tryCreateAccount(l, data, password1, password2, invitedBy)
	}

	templateInviteNewAccount.Execute(w, data)
	return data.Success
}

func tryCreateAccount(l *ldap.Conn, data *NewAccountData, pass1 string, pass2 string, invitedBy string) {
	checkFailed := false

	// Check if username is correct
	if match, err := regexp.MatchString("^[a-z0-9._-]+$", data.Username); !(err == nil && match) {
		data.ErrorInvalidUsername = true
		checkFailed = true
	}

	// Check if user exists
	userDn := config.UserNameAttr + "=" + data.Username + "," + config.UserBaseDN
	searchRq := ldap.NewSearchRequest(
		userDn,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		"(objectclass=*)",
		[]string{"dn"},
		nil)

	sr, err := l.Search(searchRq)
	if err != nil {
		data.ErrorMessage = err.Error()
		checkFailed = true
	}

	if len(sr.Entries) > 0 {
		data.ErrorUsernameTaken = true
		checkFailed = true
	}

	// Check that password is long enough
	if len(pass1) < 8 {
		data.ErrorPasswordTooShort = true
		checkFailed = true
	}

	if pass1 != pass2 {
		data.ErrorPasswordMismatch = true
		checkFailed = true
	}

	if checkFailed {
		return
	}

	// Actually create user
	req := ldap.NewAddRequest(userDn, nil)
	req.Attribute("objectclass", []string{"inetOrgPerson", "organizationalPerson", "person", "top"})
	req.Attribute("structuralobjectclass", []string{"inetOrgPerson"})
	pw, err := SSHAEncode(pass1)
	if err != nil {
		data.ErrorMessage = err.Error()
		return
	}
	req.Attribute("userpassword", []string{pw})
	req.Attribute("invitedby", []string{invitedBy})
	if len(data.DisplayName) > 0 {
		req.Attribute("displayname", []string{data.DisplayName})
	}
	if len(data.GivenName) > 0 {
		req.Attribute("givenname", []string{data.GivenName})
	}
	if len(data.Surname) > 0 {
		req.Attribute("sn", []string{data.Surname})
	}
	if len(config.InvitedMailFormat) > 0 {
		email := strings.ReplaceAll(config.InvitedMailFormat, "{}", data.Username)
		req.Attribute("mail", []string{email})
	}

	err = l.Add(req)
	if err != nil {
		data.ErrorMessage = err.Error()
		return
	}

	for _, group := range config.InvitedAutoGroups {
		req := ldap.NewModifyRequest(group, nil)
		req.Add("member", []string{userDn})
		err = l.Modify(req)
		if err != nil {
			data.WarningMessage += fmt.Sprintf("Cannot add to %s: %s\n", group, err.Error())
		}
	}

	data.Success = true
}

// ---- Code generation ----

type SendCodeData struct {
	ErrorMessage      string
	ErrorInvalidEmail bool
	Success           bool
	CodeDisplay       string
	CodeSentTo        string
	WebBaseAddress    string
}

type CodeMailFields struct {
	From           string
	To             string
	Code           string
	InviteFrom     string
	WebBaseAddress string
}

func handleInviteSendCode(w http.ResponseWriter, r *http.Request) {
	templateInviteSendCode := getTemplate("invite_send_code.html")

	login := checkInviterLogin(w, r)
	if login == nil {
		return
	}

	data := &SendCodeData{
		WebBaseAddress: config.WebAddress,
	}

	if r.Method == "POST" {
		r.ParseForm()

		choice := strings.Join(r.Form["choice"], "")
		sendto := strings.Join(r.Form["sendto"], "")

		if choice == "display" || choice == "send" {
			trySendCode(login, choice, sendto, data)
		}
	}

	templateInviteSendCode.Execute(w, data)
}

func trySendCode(login *LoginStatus, choice string, sendto string, data *SendCodeData) {
	// Generate code
	code, code_id, code_pw := genCode()

	// Create invitation object in database
	inviteDn := config.InvitationNameAttr + "=" + code_id + "," + config.InvitationBaseDN
	req := ldap.NewAddRequest(inviteDn, nil)
	pw, err := SSHAEncode(code_pw)
	if err != nil {
		data.ErrorMessage = err.Error()
		return
	}
	req.Attribute("userpassword", []string{pw})
	req.Attribute("objectclass", []string{"top", "invitationCode"})

	err = login.conn.Add(req)
	if err != nil {
		data.ErrorMessage = err.Error()
		return
	}

	// If we want to display it, do so
	if choice == "display" {
		data.Success = true
		data.CodeDisplay = code
		return
	}

	// Otherwise, we are sending a mail
	if !EMAIL_REGEXP.MatchString(sendto) {
		data.ErrorInvalidEmail = true
		return
	}

	templateMail := template.Must(template.ParseFiles(templatePath + "/invite_mail.txt"))
	buf := bytes.NewBuffer([]byte{})
	templateMail.Execute(buf, &CodeMailFields{
		To:             sendto,
		From:           config.MailFrom,
		InviteFrom:     login.WelcomeName(),
		Code:           code,
		WebBaseAddress: config.WebAddress,
	})

	log.Printf("Sending mail to: %s", sendto)
	var auth sasl.Client = nil
	if config.SMTPUsername != "" {
		auth = sasl.NewPlainClient("", config.SMTPUsername, config.SMTPPassword)
	}
	err = smtp.SendMail(config.SMTPServer, auth, config.MailFrom, []string{sendto}, buf)
	if err != nil {
		data.ErrorMessage = err.Error()
		return
	}
	log.Printf("Mail sent.")

	data.Success = true
	data.CodeSentTo = sendto
}

func genCode() (code string, code_id string, code_pw string) {
	random := make([]byte, 32)
	n, err := rand.Read(random)
	if err != nil || n != 32 {
		log.Fatalf("Could not generate random bytes: %s", err)
	}

	a := binary.BigEndian.Uint32(random[0:4])
	b := binary.BigEndian.Uint32(random[4:8])
	c := binary.BigEndian.Uint32(random[8:12])

	code = fmt.Sprintf("%03d-%03d-%03d", a%1000, b%1000, c%1000)
	code_id, code_pw = readCode(code)
	return
}

func readCode(code string) (code_id string, code_pw string) {
	// Strip everything that is not a digit
	code_digits := ""
	for _, c := range code {
		if c >= '0' && c <= '9' {
			code_digits = code_digits + string(c)
		}
	}

	id_hash := argon2.IDKey([]byte(code_digits), []byte("Guichet ID"), 2, 64*1024, 4, 32)
	pw_hash := argon2.IDKey([]byte(code_digits), []byte("Guichet PW"), 2, 64*1024, 4, 32)

	code_id = hex.EncodeToString(id_hash[:8])
	code_pw = hex.EncodeToString(pw_hash[:16])
	return
}
