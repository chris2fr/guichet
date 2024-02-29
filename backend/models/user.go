/*
Model-User Handles everything having to do with the user.
*/
package models

import (
	"fmt"
	"log"
	"strings"
	"os"
	"os/exec"

	"github.com/go-ldap/ldap/v3"

	"guichet/utils"

	"crypto/rand"
	"math/big"
)

func replaceIfContent(modifReq *ldap.ModifyRequest, key string, value string, previousValue string) error {
	if value != "" {
		modifReq.Replace(key, []string{value})
	} else if previousValue != "" {
		modifReq.Delete(key, []string{previousValue})
	}
	return nil
}

func GetUser(user User, config *ConfigFile, ldapConn *ldap.Conn) (*User, error) {
	return get(user, config, ldapConn)
}

func get(user User, config *ConfigFile, ldapConn *ldap.Conn) (*User, error) {
	searchReq := ldap.NewSearchRequest(
		user.DN,
		ldap.ScopeBaseObject,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		"(objectClass=*)",
		[]string{
			"cn",
			"givenName",
			"displayName",
			"uid",
			"sn",
			"mail",
			"description",
			"carLicense",
		},
		nil)
	searchRes, err := ldapConn.Search(searchReq)
	if err != nil {
		log.Printf("get User : %v", err)
		log.Printf("get User : %v", searchReq)
		log.Printf("get User : %v", searchRes)
		return nil, err
	}
	userEntry := searchRes.Entries[0]
	resUser := User{
		DN:           user.DN,
		GivenName:    searchRes.Entries[0].GetAttributeValue("givenName"),
		DisplayName:  searchRes.Entries[0].GetAttributeValue("displayName"),
		Description:  searchRes.Entries[0].GetAttributeValue("description"),
		SN:           searchRes.Entries[0].GetAttributeValue("sn"),
		UID:          searchRes.Entries[0].GetAttributeValue("uid"),
		CN:           searchRes.Entries[0].GetAttributeValue("cn"),
		Mail:         searchRes.Entries[0].GetAttributeValue("mail"),
		OtherMailbox: searchRes.Entries[0].GetAttributeValue("carLicense"),
		CanAdmin:     strings.EqualFold(user.DN, config.AdminAccount),
		CanInvite:    true,
		UserEntry:    userEntry,
	}
	searchReq.BaseDN = config.GroupCanAdmin
	searchReq.Filter = "(member=" + user.DN + ")"
	searchRes, err = ldapConn.Search(searchReq)
	if err == nil {
		if len(searchRes.Entries) > 0 {
			resUser.CanAdmin = true
		}
	}
	return &resUser, nil
}

func AddUser(user User, config *ConfigFile, ldapConn *ldap.Conn) error {
	return add(user, config, ldapConn)
}

// Adds a new user
func add(user User, config *ConfigFile, ldapConn *ldap.Conn) error {
	log.Printf(fmt.Sprint("Adding New User %v", user.DN))
	// LDAP Add Object
	dn := user.DN
	req := ldap.NewAddRequest(dn, nil)
	req.Attribute("objectClass", []string{"top", "person", "organizationalPerson", "inetOrgPerson"})
	if user.DisplayName != "" {
		req.Attribute("displayName", []string{user.DisplayName})
	}
	if user.GivenName != "" {
		req.Attribute("givenName", []string{user.GivenName})
	}
	if user.Mail != "" {
		req.Attribute("mail", []string{user.Mail})
	}
	if user.UID != "" {
		req.Attribute("uid", []string{user.UID})
	}
	// if user.Member != "" {
	// 	req.Attribute("member", []string{user.Member})
	// }
	if user.SN != "" {
		req.Attribute("sn", []string{user.SN})
	}
	if user.CN != "" {
		req.Attribute("cn", []string{user.CN})
	}
	if user.OtherMailbox != "" {
		req.Attribute("carLicense", []string{user.OtherMailbox})
	}
	if user.Description != "" {
		req.Attribute("description", []string{user.Description})
	}
	// Add the User
	// err := ldapConn.Add(req)
	// var ldapNewConn *ldap.Conn
	ldapNewConn, err := OpenNewUserLdap(config)
	err = ldapNewConn.Add(req)
	if err != nil {
		log.Printf(fmt.Sprintf("add(User) ldapconn.Add: %v", err))
		log.Printf(fmt.Sprintf("add(User) ldapconn.Add: %v", req))
		log.Printf(fmt.Sprintf("add(User) ldapconn.Add: %v", user))
		//return err
	} else {
		log.Printf(fmt.Sprintf("added(User) ldapconn.Add: %v", user.DN))
	}
	// passwordModifyRequest := ldap.NewPasswordModifyRequest(user.DN, "", user.Password)
	// _, err = ldapConn.PasswordModify(passwordModifyRequest)
	// if err != nil {
	// 	return err
	// }

	// Send the email

	newUserLdapConn, _ := OpenNewUserLdap(config)
	// user.OtherMailbox = ""
	err = PasswordLost(user.CN, config, newUserLdapConn)
	if err != nil {
		log.Printf("add User PasswordLost %v", err)
		log.Printf("add User PasswordLost %v", newUserLdapConn)
	}

	// sendMailTplData := SendMailTplData{
	// 	From:            "alice@resdigita.org",
	// 	To:              user.OtherMailbox,
	// 	RelTemplatePath: "user/new.email.txt",
	// 	EmailContentVars: EmailContentVarsTplData{
	// 		InviteFrom:  "alice@resdigita.org",
	// 		SendAddress: "https://www.lesgrandsvoisinss.org",
	// 		Code:        "...",
	// 	},
	// }
	// err = sendMail(sendMailTplData)
	// if err != nil {
	// 	log.Printf("add(user) sendMail: %v", err)
	// 	log.Printf("add(user) sendMail: %v", user)
	// 	log.Printf("add(user) sendMail: %v", sendMailTplData)
	// }
	// err = SyncAuthentikLDAP()
	// parts := strings.Split(user.CN,"@")
	
	// if len(parts) != 2 {
	// 	fmt.Println("Invalid email format")
	// 	// return nil
	// }

	// domain := parts[1]
	// username := parts[0]

	basePath := "/var/www/dav/data"
	// domainPath := basePath + "/" + domain
	//  := domainPath + "/" + username
	userPath := basePath + "/" + user.UID

	err = os.MkdirAll(basePath, 0775)
	// err = os.MkdirAll(domainPath, 0775)
	err = os.MkdirAll(userPath, 0775)
	// err = os.Chmod(domainPath, 0775)
	err = os.Chmod(userPath, 0775)
	if err != nil {
		log.Println("here")

			fmt.Println("Error creating folders:", err)
			// return nil
	}

	basePath = "/var/www/secret/dav"
	// domainPath = basePath + "/" + domain
	// userPath = domainPath + "/" + username
	userPath = basePath + "/" + user.UID

	err = os.MkdirAll(basePath, 0775)
	// err = os.MkdirAll(domainPath, 0775)
	err = os.MkdirAll(userPath, 0775)
	// err = os.Chmod(domainPath, 0775)
	err = os.Chmod(userPath, 0775)
	if err != nil {
		log.Println("here")

			fmt.Println("Error creating folders:", err)
			// return nil
	}

	log.Println("here")
	fmt.Println("Folders created successfully:", userPath)


	err = utils.CopyFiles(basePath + "/example.com/templateuser", userPath )
	if err != nil {
		log.Println("here")
		fmt.Println("Error copying files:", err)
	}

	log.Println("here")
	fmt.Println("Files /example.com/templateuser copied successfully")

	// Create user for file browser

	cmd := exec.Command(fmt.Sprintf("filebrowser","user","add", user.UID, RandPW(16)))
	cmd.Dir = "/home/filebrowser/"
  err = cmd.Run()

  if err != nil {
		log.Println("here")
		fmt.Println("Error creating user %v :", err)
  } 

	return err
}

func RandPW (length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

	randomString := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := range randomString {
		randomByte, _ := rand.Int(rand.Reader, charsetLen)
		randomString[i] = charset[randomByte.Int64()]
	}

	return string(randomString)
}

func ModifyUser(user User, config *ConfigFile, ldapConn *ldap.Conn) error {
	return modify(user, config, ldapConn)
}

func modify(user User, config *ConfigFile, ldapConn *ldap.Conn) error {
	modify_request := ldap.NewModifyRequest(user.DN, nil)
	previousUser, err := get(user, config, ldapConn)
	if err != nil {
		return err
	}
	replaceIfContent(modify_request, "displayName", user.DisplayName, previousUser.DisplayName)
	replaceIfContent(modify_request, "givenName", user.GivenName, previousUser.GivenName)
	replaceIfContent(modify_request, "sn", user.SN, previousUser.SN)
	// replaceIfContent(modify_request, "carLicense", user.OtherMailbox, user.OtherMailbox)
	replaceIfContent(modify_request, "description", user.Description, previousUser.Description)
	err = ldapConn.Modify(modify_request)
	if err != nil {
		log.Printf(fmt.Sprintf("71: %v", err))
		log.Printf(fmt.Sprintf("72: %v", modify_request))
		log.Printf(fmt.Sprintf("73: %v", user))
		return err
	}
	// err = SyncAuthentikLDAP()

	return err
}

func PassWD(user User, config *ConfigFile, ldapConn *ldap.Conn) error {
	passwordModifyRequest := ldap.NewPasswordModifyRequest(user.DN, "", user.Password)
	// log.Printf(fmt.Sprintf("model-user PassWD : %v : %v", user.DN, user.Password))
	_, err := ldapConn.PasswordModify(passwordModifyRequest)
	if err != nil {
		log.Printf(fmt.Sprintf("model-user PassWD : %v %v", err, ldapConn))
		log.Printf(fmt.Sprintf("model-user PassWD : %v", user))
	}
	// err = SyncAuthentikLDAP()
	return err
}

func Bind(user User, config *ConfigFile, ldapConn *ldap.Conn) error {
	return ldapConn.Bind(user.DN, user.Password)
}
