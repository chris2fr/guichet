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

	// "github.com/emersion/go-sasl"
	// "github.com/emersion/go-smtp"
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

	// if !login.CanInvite {
	// 	http.Error(w, "Not authorized to invite new users.", http.StatusUnauthorized)
	// 	return nil
	// }

	return login
}

// New account creation directly from interface

func handleInviteNewAccount(w http.ResponseWriter, r *http.Request) {
	// l := ldapOpen(w)
	// l.Bind(config.NewUserDN, config.NewUserPassword)

	login := checkInviterLogin(w, r)
	if login == nil {
		return
	}
	// l, _ := ldap.DialURL(config.LdapServerAddr)
	// l.Bind(config.NewUserDN, config.NewUserPassword)
	handleNewAccount(w, r, login.conn, config.NewUserDN)
}

// New account creation using code

func handleInvitationCode(w http.ResponseWriter, r *http.Request) {
	code := mux.Vars(r)["code"]
	code_id, code_pw := readCode(code)

	// log.Printf(code_pw)

	login := checkLogin(w, r)

	// l := ldapOpen(w)
	// if l == nil {
	// 	return
	// }

	inviteDn := config.InvitationNameAttr + "=" + code_id + "," + config.InvitationBaseDN
	err := login.conn.Bind(inviteDn, code_pw)
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
	sr, err := login.conn.Search(sReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(sr.Entries) != 1 {
		http.Error(w, fmt.Sprintf("Expected 1 entry, got %d", len(sr.Entries)), http.StatusInternalServerError)
		return
	}

	invitedBy := sr.Entries[0].GetAttributeValue("creatorsname")

	if handleNewAccount(w, r, login.conn, invitedBy) {
		del_req := ldap.NewDelRequest(inviteDn, nil)
		err = login.conn.Del(del_req)
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
	Mail        string
	SuggestPW   string

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

		newUser := NewUser{}
		// login := checkLogin(w, r)

		newUser.CN = fmt.Sprintf("%s@%s", strings.TrimSpace(strings.Join(r.Form["username"], "")), "lesgv.com")
		newUser.DisplayName = strings.TrimSpace(strings.Join(r.Form["displayname"], ""))
		newUser.GivenName = strings.TrimSpace(strings.Join(r.Form["givenname"], ""))
		newUser.SN = strings.TrimSpace(strings.Join(r.Form["surname"], ""))
		newUser.UID = strings.TrimSpace(strings.Join(r.Form["username"], ""))
		newUser.Mail = strings.TrimSpace(strings.Join(r.Form["mail"], ""))
		newUser.DN = "cn=" + newUser.CN + "," + config.InvitationBaseDN

		password1 := strings.Join(r.Form["password"], "")
		password2 := strings.Join(r.Form["password2"], "")

		if password1 != password2 {
			data.Success = false
			data.ErrorPasswordMismatch = true
		} else {
			newUser.Password = password2
			data.Success = addNewUser(newUser, config, l)
			http.Redirect(w, r, "/admin/ldap/"+newUser.DN, http.StatusFound)
		}

		// tryCreateAccount(l, data, password1, password2, invitedBy)

	} else {
		data.SuggestPW = fmt.Sprintf("%s", suggestPassword())
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

	// carLicense

	if r.Method == "POST" {
		r.ParseForm()
		data := &SendCodeData{
			WebBaseAddress: config.WebAddress,
		}

		// modify_request := ldap.NewModifyRequest(login.UserEntry.DN, nil)
		// // choice := strings.Join(r.Form["choice"], "")
		// // sendto := strings.Join(r.Form["sendto"], "")
		code, code_id, code_pw := genCode()
		log.Printf(fmt.Sprintf("272: %v %v %v", code, code_id, code_pw))
		// // Create invitation object in database
		// modify_request.Add("carLicense", []string{fmt.Sprintf("%s,%s,%s",code, code_id, code_pw)})
		// err := login.conn.Modify(modify_request)
		// if err != nil {
		// 	data.ErrorMessage = err.Error()
		// 	// return
		// } else {
		// 	data.Success = true
		// 	data.CodeDisplay = code
		// }
		log.Printf(fmt.Sprintf("279: %v %v %v", code, code_id, code_pw))
		addReq := ldap.NewAddRequest("documentIdentifier="+code_id+","+config.InvitationBaseDN, nil)
		addReq.Attribute("objectClass", []string{"top", "document", "simpleSecurityObject"})
		addReq.Attribute("cn", []string{code})
		addReq.Attribute("userPassword", []string{code_pw})
		addReq.Attribute("documentIdentifier", []string{code_id})
		log.Printf(fmt.Sprintf("285: %v %v %v", code, code_id, code_pw))
		log.Printf(fmt.Sprintf("286: %v", addReq))
		err := login.conn.Add(addReq)
		if err != nil {
			data.ErrorMessage = err.Error()
			// return
		} else {
			data.Success = true
			data.CodeDisplay = code
		}

		templateInviteSendCode.Execute(w, data)

		// if choice == "display" || choice == "send" {
		// 	log.Printf(fmt.Sprintf("260: %v %v %v %v", login, choice, sendto, data))
		// 	trySendCode(login, choice, sendto, data)
		// }
	}

}

func trySendCode(login *LoginStatus, choice string, sendto string, data *SendCodeData) {
	log.Printf(fmt.Sprintf("269: %v %v %v %v", login, choice, sendto, data))
	// Generate code
	code, code_id, code_pw := genCode()
	log.Printf(fmt.Sprintf("272: %v %v %v", code, code_id, code_pw))
	// Create invitation object in database

	// len_base_dn := len(strings.Split(config.BaseDN, ","))
	// dn_split := strings.Split(super_dn, ",")
	// for i := len_base_dn + 1; i <= len(dn_split); i++ {
	// 	path = append(path, PathItem{
	// 		DN:         strings.Join(dn_split[len(dn_split)-i:len(dn_split)], ","),
	// 		Identifier: dn_split[len(dn_split)-i],
	// 	})
	// }
	// data := &SendCodeData{
	// 	WebBaseAddress: config.WebAddress,
	// }
	// // Handle data
	// data := &CreateData{
	// 	SuperDN: super_dn,
	// 	Path:    path,
	// }
	// data.IdType = config.UserNameAttr
	// data.StructuralObjectClass = "inetOrgPerson"
	// data.ObjectClass = "inetOrgPerson\norganizationalPerson\nperson\ntop"
	// data.IdValue = strings.TrimSpace(strings.Join(r.Form["idvalue"], ""))
	// data.DisplayName = strings.TrimSpace(strings.Join(r.Form["displayname"], ""))
	// data.GivenName = strings.TrimSpace(strings.Join(r.Form["givenname"], ""))
	// data.Mail = strings.TrimSpace(strings.Join(r.Form["mail"], ""))
	// data.Member = strings.TrimSpace(strings.Join(r.Form["member"], ""))
	// data.Description = strings.TrimSpace(strings.Join(r.Form["description"], ""))
	// data.SN = strings.TrimSpace(strings.Join(r.Form["sn"], ""))
	// object_class := []string{}
	// for _, oc := range strings.Split(data.ObjectClass, "\n") {
	// 	x := strings.TrimSpace(oc)
	// 	if x != "" {
	// 		object_class = append(object_class, x)
	// 	}
	// }
	// dn := data.IdType + "=" + data.IdValue + "," + super_dn
	// 		req := ldap.NewAddRequest(dn, nil)
	// 		req.Attribute("objectclass", object_class)
	// 		// req.Attribute("mail", []string{data.IdValue})
	//     /*
	// 		if data.StructuralObjectClass != "" {
	// 			req.Attribute("structuralobjectclass", []string{data.StructuralObjectClass})
	// 		}
	//     */
	// 		if data.DisplayName != "" {
	// 			req.Attribute("displayname", []string{data.DisplayName})
	// 		}
	// 		if data.GivenName != "" {
	// 			req.Attribute("givenname", []string{data.GivenName})
	// 		}
	// 		if data.Mail != "" {
	// 			req.Attribute("mail", []string{data.Mail})
	// 		}
	// 		if data.Member != "" {
	// 			req.Attribute("member", []string{data.Member})
	// 		}
	// 		if data.SN != "" {
	// 			req.Attribute("sn", []string{data.SN})
	// 		}
	// 		if data.Description != "" {
	// 			req.Attribute("description", []string{data.Description})
	// 		}
	// 		err := login.conn.Add(req)
	//     // log.Printf(fmt.Sprintf("899: %v",err))
	//     // log.Printf(fmt.Sprintf("899: %v",req))
	//     // log.Printf(fmt.Sprintf("899: %v",data))
	// 		if err != nil {
	// 			data.Error = err.Error()
	// 		} else {
	// 			if template == "ml" {
	// 				http.Redirect(w, r, "/admin/mailing/"+data.IdValue, http.StatusFound)
	// 			} else {
	// 				http.Redirect(w, r, "/admin/ldap/"+dn, http.StatusFound)
	// 			}
	// 		}

	// inviteDn := config.InvitationNameAttr + "=" + code_id + "," + config.InvitationBaseDN
	// req := ldap.NewAddRequest(inviteDn, nil)
	// pw, err := SSHAEncode(code_pw)
	// if err != nil {
	// 	data.ErrorMessage = err.Error()
	// 	return
	// }
	// req.Attribute("employeeNumber", []string{pw})
	// req.Attribute("objectclass", []string{"top", "invitationCode"})

	// err = login.conn.Add(req)
	// if err != nil {
	// 	log.Printf(fmt.Sprintf("286: %v", req))
	// 	data.ErrorMessage = err.Error()
	// 	return
	// }

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
	// var auth sasl.Client = nil
	// if config.SMTPUsername != "" {
	// 	auth = sasl.NewPlainClient("", config.SMTPUsername, config.SMTPPassword)
	// }
	// err = smtp.SendMail(config.SMTPServer, auth, config.MailFrom, []string{sendto}, buf)
	// if err != nil {
	// 	data.ErrorMessage = err.Error()
	// 	return
	// }
	// log.Printf("Mail sent.")

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
	log.Printf(fmt.Sprintf("342: %v %v %v", code, code_id, code_pw))
	return code, code_id, code_pw
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
	return code_id, code_pw
}
