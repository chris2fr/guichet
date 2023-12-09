/*
Creates the webpages to be processed by Guichet
*/
package views

import (
	"crypto/tls"
	"encoding/json"
	"guichet/models"
	"io/ioutil"
	"net"

	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	// "net/http"
	"strings"

	"github.com/go-ldap/ldap/v3"
	"github.com/gorilla/sessions"
)

const SESSION_NAME = "guichet_session"

var templatePath = "./templates"
var GuichetSessionStore sessions.Store = nil

type EntryList []*ldap.Entry
type LoginInfo struct {
	Username string
	DN       string
	Password string
}
func ReadConfig() models.ConfigFile {
	// Default configuration values for certain fields
	flag.Parse()
	var configFlag = flag.String("config", "./config.json", "Configuration file path")

	config_file := models.ConfigFile{
		HttpBindAddr:   ":9992",
		LdapServerAddr: "ldap://127.0.0.1:389",

		UserNameAttr:  "uid",
		GroupNameAttr: "gid",

		InvitationNameAttr: "cn",
		InvitedAutoGroups:  []string{},

		Org: "ResDigita",
	}

	_, err := os.Stat(*configFlag)
	if os.IsNotExist(err) {
		log.Fatalf("Could not find Guichet configuration file at %s. Please create this file, for exemple starting with config.json.exemple and customizing it for your deployment.", *configFlag)
	}

	if err != nil {
		log.Fatal(err)
	}

	bytes, err := ioutil.ReadFile(*configFlag)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(bytes, &config_file)
	if err != nil {
		log.Fatal(err)
	}

	return config_file
}
type LoginStatus struct {
	Info      *LoginInfo
	conn      *ldap.Conn
	UserEntry *ldap.Entry
	Common    NestedCommonTplData
}
type NestedCommonTplData struct {
	Error          string
	ErrorMessage   string
	CanAdmin       bool
	CanInvite      bool
	LoggedIn       bool
	Success        bool
	WarningMessage string
	WebsiteName    string
	WebsiteURL     string
}
type CodeMailFields struct {
	From           string
	To             string
	Code           string
	InviteFrom     string
	WebBaseAddress string
	Common         NestedCommonTplData
}

var config = ReadConfig()

func ldapOpen(w http.ResponseWriter) (*ldap.Conn, error) {
	if config.LdapTLS {
		tlsConf := &tls.Config{
			ServerName:         config.LdapServerAddr,
			InsecureSkipVerify: true,
		}
		return ldap.DialTLS("tcp", net.JoinHostPort(config.LdapServerAddr, "636"), tlsConf)
	} else {
		return ldap.DialURL("ldap://" + config.LdapServerAddr)
	}

	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	log.Printf(fmt.Sprintf("27: %v %v", err, l))
	// 	return nil
	// }

	// return l
}


// type keyView struct {
// 	Status *LoginStatus
// 	Key    *garage.KeyInfo
// }
// type webInspectView struct {
// 	Status      *LoginStatus
// 	Key         *garage.KeyInfo
// 	Bucket      *garage.BucketInfo
// 	IndexDoc    string
// 	ErrorDoc    string
// 	MaxObjects  int64
// 	MaxSize     int64
// 	UsedSizePct float64
// }
// type webListView struct {
// 	Status *LoginStatus
// 	Key    *garage.KeyInfo
// }
type LayoutTemplateData struct {
	Common NestedCommonTplData
	Login  NestedLoginTplData
	Data   any
}
type NestedLoginTplData struct {
	Login    *LoginStatus
	Username string
	Status   *LoginStatus
}


func execTemplate(w http.ResponseWriter, t *template.Template, commonData NestedCommonTplData, loginData NestedLoginTplData, data any) error {
	commonData.WebsiteURL = config.WebAddress
	commonData.WebsiteName = config.Org
	return t.Execute(w, LayoutTemplateData{
		Common: commonData,
		Login:  loginData,
		Data:   data,
	})
}


func (login *LoginStatus) WelcomeName() string {
	ret := login.UserEntry.GetAttributeValue("givenName")
	if ret == "" {
		ret = login.UserEntry.GetAttributeValue("displayName")
	}
	if ret == "" {
		ret = login.Info.Username
	}
	return ret
}


type AdminUsersTplData struct {
	UserNameAttr string
	UserBaseDN   string
	Users        EntryList
	Common       NestedCommonTplData
	Login        NestedLoginTplData
}
type AdminLDAPTplData struct {
	DN string

	Path          []PathItem
	ChildrenOU    []Child
	ChildrenOther []Child
	CanAddChild   bool
	Props         map[string]*PropValues
	CanDelete     bool

	HasMembers         bool
	Members            []EntryName
	PossibleNewMembers []EntryName
	HasGroups          bool
	Groups             []EntryName
	PossibleNewGroups  []EntryName

	ListMemGro map[string]string

	Common NestedCommonTplData
	Login  NestedLoginTplData
}
type AdminMailingListTplData struct {
	Common             NestedCommonTplData
	Login              NestedLoginTplData
	MailingNameAttr    string
	MailingBaseDN      string
	MailingList        *ldap.Entry
	Members            EntryList
	PossibleNewMembers EntryList
	AllowGuest         bool
}
type AdminMailingTplData struct {
	Common          NestedCommonTplData
	Login           NestedLoginTplData
	MailingNameAttr string
	MailingBaseDN   string
	MailingLists    EntryList
}
type AdminGroupsTplData struct {
	Common        NestedCommonTplData
	Login         NestedLoginTplData
	GroupNameAttr string
	GroupBaseDN   string
	Groups        EntryList
}
type EntryName struct {
	DN   string
	Name string
}
type Child struct {
	DN         string
	Identifier string
	Name       string
}
type PathItem struct {
	DN         string
	Identifier string
	Active     bool
}
type PropValues struct {
	Name      string
	Values    []string
	Editable  bool
	Deletable bool
}
type CreateData struct {
	SuperDN  string
	Path     []PathItem
	Template string

	IdType                string
	IdValue               string
	DisplayName           string
	GivenName             string
	Member                string
	Mail                  string
	Description           string
	StructuralObjectClass string
	ObjectClass           string
	SN                    string
	OtherMailbox          string

	Common NestedCommonTplData
	Login  NestedLoginTplData
}


type HomePageData struct {
	Common NestedCommonTplData
	Login  NestedLoginTplData
	BaseDN string
	Org    string
}
type PasswordFoundData struct {
	Common       NestedCommonTplData
	Login        NestedLoginTplData
	Username     string
	Mail         string
	OtherMailbox string
}
type PasswordLostData struct {
	Common       NestedCommonTplData
	ErrorMessage string
	Success      bool
	Username     string
	Mail         string
	OtherMailbox string
}
type NewAccountData struct {
	Username     string
	DisplayName  string
	GivenName    string
	Surname      string
	Mail         string
	SuggestPW    string
	OtherMailbox string

	ErrorUsernameTaken    bool
	ErrorInvalidUsername  bool
	ErrorPasswordTooShort bool
	ErrorPasswordMismatch bool
	Common                NestedCommonTplData
	NewUserDefaultDomain  string
}
type SendCodeData struct {
	Common            NestedCommonTplData
	ErrorInvalidEmail bool

	CodeDisplay    string
	CodeSentTo     string
	WebBaseAddress string
}


type ProfileTplData struct {
	Mail         string
	MailValues   []string
	DisplayName  string
	GivenName    string
	Surname      string
	Description  string
	OtherMailbox string
	Common       NestedCommonTplData
	Login        NestedLoginTplData
}

//ProfilePicture string
//Visibility     string

type PasswdTplData struct {
	Common        NestedCommonTplData
	Login         NestedLoginTplData
	TooShortError bool
	NoMatchError  bool
}


type LoginFormData struct {
	Username  string
	WrongUser bool
	WrongPass bool
	Common    NestedCommonTplData
}



type WrapperTemplate struct {
	Template *template.Template
}



func getTemplate(name string) *template.Template {
	return template.Must(template.New("layout.html").Funcs(template.FuncMap{
		"contains": strings.Contains,
	}).ParseFiles(
		templatePath+"/layout.html",
		templatePath+"/"+name,
	))
}



