package views

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"guichet/models"
	"guichet/utils"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strings"

	// "crypto/rand"

	// "github.com/emersion/go-smtp"
	"github.com/go-ldap/ldap/v3"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/argon2"

	"github.com/dchest/captcha"
	// "flag"

	// "github.com/mojocn/base64Captcha"
	// "math/rand"
	"time"
	"errors"
	// "encoding/json"
	// "clevergo.tech/captchas/stores/memstore"


)

var EMAIL_REGEXP = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

var captchaStore = captcha.NewMemoryStore(101, 10 * time.Minute)




const (
	// Default number of digits in captcha solution.
	DefaultLen = 6
	// The number of captchas created that triggers garbage collection used
	// by default store.
	CollectNum = 100
	// Expiration time of captchas used by default store.
	Expiration = 10 * time.Minute
)

const (
	// Standard width and height of a captcha image.
	StdWidth  = 240
	StdHeight = 80
)

var (
	ErrNotFound = errors.New("captcha: id not found")
)

// func init() {
// 	//init rand seed
// 	rand.Seed(time.Now().UnixNano())
// }
// var (
// 	flagImage = flag.Bool("i", true, "output image captcha")
// 	flagAudio = flag.Bool("a", false, "output audio captcha")
// 	flagLang  = flag.String("lang", "en", "language for audio captcha")
// 	flagLen   = flag.Int("len", captcha.DefaultLen, "length of captcha")
// 	flagImgW  = flag.Int("width", captcha.StdWidth, "image captcha width")
// 	flagImgH  = flag.Int("height", captcha.StdHeight, "image captcha height")
// )

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



func HandleInviteNewAccount(w http.ResponseWriter, r *http.Request) {
	l, err := ldapOpen(w)
	if err != nil {
		log.Printf("view-invite.go - HandleInviteNewAccount - ldapOpen : %v", err)
		log.Printf("view-invite.go - HandleInviteNewAccount - ldapOpen: %v", l)
	}
	if l == nil {
		return
	}

	err = l.Bind(config.NewUserDN, config.NewUserPassword)
	if err != nil {
		log.Printf("view-invite.go - HandleInviteNewAccount - l.Bind : %v", err)
		log.Printf("view-invite.go - HandleInviteNewAccount - l.Bind: %v", config.NewUserDN)
		panic(fmt.Sprintf("view-invite.go - HandleInviteNewAccount - l.Bind : %v", err))
	}
	HandleNewAccount(w, r, l, config.NewUserDN)
}

// New account creation using code
func HandleInvitationCode(w http.ResponseWriter, r *http.Request) {
	code := mux.Vars(r)["code"]
	code_id, code_pw := readCode(code)
	login := checkLogin(w, r)
	inviteDn := config.InvitationNameAttr + "=" + code_id + "," + config.InvitationBaseDN
	err := login.conn.Bind(inviteDn, code_pw)
	if err != nil {
		templateInviteInvalidCode := getTemplate("user/code/invalid.html")
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
	if HandleNewAccount(w, r, login.conn, invitedBy) {
		del_req := ldap.NewDelRequest(inviteDn, nil)
		err = login.conn.Del(del_req)
		if err != nil {
			log.Printf("Could not delete invitation %s: %s", inviteDn, err)
		}
	}
}

// type Store interface {
// 	// Set sets the digits for the captcha id.
// 	Set(id string, value string)

// 	// Get returns stored digits for the captcha id. Clear indicates
// 	// whether the captcha must be deleted from the store.
// 	Get(id string, clear bool) string
	
//     //Verify captcha's answer directly
// 	Verify(id, answer string, clear bool) bool
// }

// // Driver captcha interface for captcha engine to to write staff
// type Driver interface {
// 	//DrawCaptcha draws binary item
// 	DrawCaptcha(content string) (item Item, err error)
// 	//GenerateIdQuestionAnswer creates rand id, content and answer
// 	GenerateIdQuestionAnswer() (id, q, a string)
// }


// // Captcha captcha basic information.
// type Captcha struct {
// 	Driver Driver
// 	Store  Store
// }


// //NewCaptcha creates a captcha instance from driver and store
// func NewCaptcha(driver Driver, store Store) *Captcha {
// 	return &Captcha{Driver: driver, Store: store}
// }

// //Generate generates a random id, base64 image string or an error if any
// func (c *Captcha) Generate() (id, b64s string, err error) {
// 	id,content, answer := c.Driver.GenerateIdQuestionAnswer()
// 	item, err := c.Driver.DrawCaptcha(content)
// 	if err != nil {
// 		return "", "", err
// 	}
// 	c.Store.Set(id, answer)
// 	b64s = item.EncodeB64string()
// 	return
// }


// //Verify by a given id key and remove the captcha value in store,
// //return boolean value.
// //if you has multiple captcha instances which share a same store.
// //You may want to call `store.Verify` method instead.
// func (c *Captcha) Verify(id, answer string, clear bool) (match bool) {
// 	match = c.Store.Get(id, clear) == answer
// 	return
// }

// func (c *Captcha) Generate() (id, b64s string, err error) {
// 	id,content, answer := c.Driver.GenerateIdQuestionAnswer()
// 	item, err := c.Driver.DrawCaptcha(content)
// 	if err != nil {
// 		return "", "", err
// 	}
// 	c.Store.Set(id, answer)
// 	b64s = item.EncodeB64string()
// 	return
// }

// //if you has multiple captcha instances which shares a same store. You may want to use `store.Verify` method instead.
// //Verify by given id key and remove the captcha value in store, return boolean value.
// func (c *Captcha) Verify(id, answer string, clear bool) (match bool) {
// 	match = c.Store.Get(id, clear) == answer
// 	return
// }


// func GenerateCaptcha() string {
// 	driver := base64Captcha.NewDriverDigit(100, 240, 4, 0.7, 80)
// 	captcha := base64Captcha.NewCaptcha(driver, base64Captcha.DefaultMemStore)
// 	id, b64s, _ := captcha.Generate()
// 	return b64s
// }

// func VerifyCaptcha(c *context) {
// 	inputValue := c.PostForm("captcha")
// 	captchaId := c.PostForm("captcha_id")
// 	isValid := base64Captcha.VerifyCaptcha(captchaId, inputValue)

// 	if isValid {
// 		c.String(200, "CAPTCHA is valid.")
// 	} else {
// 		c.String(400, "CAPTCHA is invalid.")
// 	}
// }




// type configJsonBody struct {
// 	Id            string
// 	CaptchaType   string
// 	VerifyValue   string
// 	DriverAudio   *base64Captcha.DriverAudio
// 	DriverString  *base64Captcha.DriverString
// 	DriverChinese *base64Captcha.DriverChinese
// 	DriverMath    *base64Captcha.DriverMath
// 	DriverDigit   *base64Captcha.DriverDigit
// }

// var store = base64Captcha.DefaultMemStore

// // base64Captcha create http handler
// func GenerateCaptchaHandler(w http.ResponseWriter, r *http.Request) {
// 	//parse request parameters
// 	decoder := json.NewDecoder(r.Body)
// 	var param configJsonBody
// 	err := decoder.Decode(&param)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	defer r.Body.Close()
// 	var driver base64Captcha.Driver

// 	//create base64 encoding captcha
// 	switch param.CaptchaType {
// 	case "audio":
// 		driver = param.DriverAudio
// 	case "string":
// 		driver = param.DriverString.ConvertFonts()
// 	case "math":
// 		driver = param.DriverMath.ConvertFonts()
// 	case "chinese":
// 		driver = param.DriverChinese.ConvertFonts()
// 	default:
// 		driver = param.DriverDigit
// 	}
// 	c := base64Captcha.NewCaptcha(driver, store)
// 	id, b64s, msg, err := c.Generate()
// 	body := map[string]interface{}{"code": 1, "data": b64s, "captchaId": id, "msg": msg}
// 	if err != nil {
// 		body = map[string]interface{}{"code": 0, "msg": err.Error()}
// 	}
// 	w.Header().Set("Content-Type", "application/json; charset=utf-8")
// 	json.NewEncoder(w).Encode(body)
// }

// // base64Captcha verify http handler
// func CaptchaVerifyHandle(w http.ResponseWriter, r *http.Request) {

// 	//parse request json body
// 	decoder := json.NewDecoder(r.Body)
// 	var param configJsonBody
// 	err := decoder.Decode(&param)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	defer r.Body.Close()
// 	//verify the captcha
// 	body := map[string]interface{}{"code": 0, "msg": "failed"}
// 	if store.Verify(param.Id, param.VerifyValue, true) {
// 		body = map[string]interface{}{"code": 1, "msg": "ok"}
// 	}

// 	//set json response
// 	w.Header().Set("Content-Type", "application/json; charset=utf-8")

// 	json.NewEncoder(w).Encode(body)
// }




// Common functions for new account
func HandleNewAccount(w http.ResponseWriter, r *http.Request, l *ldap.Conn, invitedBy string) bool {
	templateInviteNewAccount := getTemplate("user/new.html")
	captcha.SetCustomStore(captchaStore)


	data := NewAccountData{
		NewUserDefaultDomain: config.NewUserDefaultDomain,
		// CaptchaId: captcha.New(),
	}
	if r.Method != "POST" {
		data.CaptchaId = captcha.New()
	}
		
	if r.Method == "POST" {
		r.ParseForm()
		// if !captcha.VerifyString(r.FormValue("captchaId"), r.FormValue("captchaSolution")) {
		// 	data.Common.Success = false
		// 	data.Common.ErrorMessage = "Captcha KO"
		// 	return false
		// }
		newUser := models.User{}
		newUser.DisplayName = strings.TrimSpace(strings.Join(r.Form["displayname"], ""))
		newUser.GivenName = strings.TrimSpace(strings.Join(r.Form["givenname"], ""))
		newUser.SN = strings.TrimSpace(strings.Join(r.Form["surname"], ""))
		newUser.OtherMailbox = strings.TrimSpace(strings.Join(r.Form["othermailbox"], ""))
		newUser.Mail = strings.TrimSpace(strings.Join(r.Form["mail"], ""))
		// newUser.UID = strings.TrimSpace(strings.Join(r.Form["username"], ""))
		// newUser.CN = strings.TrimSpace(strings.Join(r.Form["username"], ""))
		newUser.UID = strings.TrimSpace(strings.Join(r.Form["mail"], ""))
		newUser.CN = strings.TrimSpace(strings.Join(r.Form["mail"], ""))
		newUser.DN = "cn=" + strings.TrimSpace(strings.Join(r.Form["username"], "")) + "," + config.UserBaseDN
		password1 := strings.Join(r.Form["password"], "")
		password2 := strings.Join(r.Form["password2"], "")
		if password1 != password2 {
			data.Common.Success = false
			data.ErrorPasswordMismatch = true
		} else {
			newUser.Password = password2
			l.Bind(config.NewUserDN, config.NewUserPassword)
			err := models.AddUser(newUser, &config, l)
			if err != nil {
				data.Common.Success = false
				data.Common.ErrorMessage = err.Error()
			}
			http.Redirect(w, r, "/user/wait", http.StatusFound)
		}
		// tryCreateAccount(l, data, password1, password2, invitedBy)
	} else {
		data.SuggestPW = fmt.Sprintf("%s", models.SuggestPassword())
	}
	data.Common.CanAdmin = false
	data.Common.LoggedIn = false

	data.Common.Success = captcha.VerifyString(r.FormValue("captchaId"), r.FormValue("captchaSolution"))
	data.Common.ErrorMessage = r.FormValue("captchaSolution")

	templateInviteNewAccount.Execute(w, data)
	l.Close()
	return data.Common.Success
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
		data.Common.ErrorMessage = err.Error()
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
	pw, err := utils.SSHAEncode(pass1)
	if err != nil {
		data.Common.ErrorMessage = err.Error()
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
		data.Common.ErrorMessage = err.Error()
		return
	}
	for _, group := range config.InvitedAutoGroups {
		req := ldap.NewModifyRequest(group, nil)
		req.Add("member", []string{userDn})
		err = l.Modify(req)
		if err != nil {
			data.Common.WarningMessage += fmt.Sprintf("Cannot add to %s: %s\n", group, err.Error())
		}
	}
	data.Common.Success = true
}

// ---- Code generation ----
func HandleInviteSendCode(w http.ResponseWriter, r *http.Request) {
	templateInviteSendCode := getTemplate("user/code/send.html")
	login := checkInviterLogin(w, r)
	if login == nil {
		return
	}
	if r.Method == "POST" {
		r.ParseForm()
		data := &SendCodeData{
			WebBaseAddress: config.WebAddress,
		}
		// modify_request := ldap.NewModifyRequest(login.UserEntry.DN, nil)
		// // choice := strings.Join(r.Form["choice"], "")
		// // sendto := strings.Join(r.Form["sendto"], "")
		code, code_id, code_pw := genCode()
		log.Printf("272: %v %v %v", code, code_id, code_pw)
		// // Create invitation object in database
		// modify_request.Add("carLicense", []string{fmt.Sprintf("%s,%s,%s",code, code_id, code_pw)})
		// err := login.conn.Modify(modify_request)
		// if err != nil {
		// 	data.Common.ErrorMessage = err.Error()
		// 	// return
		// } else {
		// 	data.Common.Success = true
		// 	data.CodeDisplay = code
		// }
		log.Printf("279: %v %v %v", code, code_id, code_pw)
		addReq := ldap.NewAddRequest("documentIdentifier="+code_id+","+config.InvitationBaseDN, nil)
		addReq.Attribute("objectClass", []string{"top", "document", "simpleSecurityObject"})
		addReq.Attribute("cn", []string{code})
		addReq.Attribute("userPassword", []string{code_pw})
		addReq.Attribute("documentIdentifier", []string{code_id})
		log.Printf("285: %v %v %v", code, code_id, code_pw)
		log.Printf("286: %v", addReq)
		err := login.conn.Add(addReq)
		if err != nil {
			data.Common.ErrorMessage = err.Error()
			// return
		} else {
			data.Common.Success = true
			data.CodeDisplay = code
		}
		data.Common.CanAdmin = login.Common.CanAdmin
		templateInviteSendCode.Execute(w, data)
		// if choice == "display" || choice == "send" {
		// 	log.Printf("260: %v %v %v %v", login, choice, sendto, data)
		// 	trySendCode(login, choice, sendto, data)
		// }
	}
	login.conn.Close()

}

func trySendCode(login *LoginStatus, choice string, sendto string, data *SendCodeData) {
	log.Printf("269: %v %v %v %v", login, choice, sendto, data)
	// Generate code
	code, code_id, code_pw := genCode()
	log.Printf("272: %v %v %v", code, code_id, code_pw)
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
	//     // log.Printf("899: %v",err)
	//     // log.Printf("899: %v",req)
	//     // log.Printf("899: %v",data)
	// 		if err != nil {
	// 			data.Common.Error = err.Error()
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
	// 	data.Common.ErrorMessage = err.Error()
	// 	return
	// }
	// req.Attribute("employeeNumber", []string{pw})
	// req.Attribute("objectclass", []string{"top", "invitationCode"})
	// err = login.conn.Add(req)
	// if err != nil {
	// 	log.Printf("286: %v", req)
	// 	data.Common.ErrorMessage = err.Error()
	// 	return
	// }

	// If we want to display it, do so
	if choice == "display" {
		data.Common.Success = true
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
	templateMail.Execute(buf, &models.CodeMailFields{
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
	// 	data.Common.ErrorMessage = err.Error()
	// 	return
	// }
	// log.Printf("Mail sent.")
	data.Common.Success = true
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
	log.Printf("342: %v %v %v", code, code_id, code_pw)
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
