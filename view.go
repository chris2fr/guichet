/*
Creates the webpages to be processed by Guichet
*/
package main

import (
	"html/template"
	"net/http"

	// "net/http"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

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
type NestedLoginTplData struct {
	Login    *LoginStatus
	Username string
	Status   *LoginStatus
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
type SearchResult struct {
	DN             string
	Id             string
	DisplayName    string
	Email          string
	Description    string
	ProfilePicture string
}
type SearchResults struct {
	Results []SearchResult
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

type CodeMailFields struct {
	From           string
	To             string
	Code           string
	InviteFrom     string
	WebBaseAddress string
	Common         NestedCommonTplData
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
type LoginInfo struct {
	Username string
	DN       string
	Password string
}
type LoginStatus struct {
	Info      *LoginInfo
	conn      *ldap.Conn
	UserEntry *ldap.Entry
	Common    NestedCommonTplData
}
type LoginFormData struct {
	Username  string
	WrongUser bool
	WrongPass bool
	Common    NestedCommonTplData
}

type EmailContentVarsTplData struct {
	Code        string
	SendAddress string
	InviteFrom  string
}

// Data to be passed to an email for sending
type SendMailTplData struct {
	// Sender of the email
	To string
	// Receiver of the email
	From string
	// Relative path (without leading /) to the email template in the templates folder
	// usually ending in .txt
	RelTemplatePath string
	// Variables to be included in the template of the email
	EmailContentVars EmailContentVarsTplData
}

type WrapperTemplate struct {
	Template *template.Template
}

var templatePath = "./templates"

func getTemplate(name string) *template.Template {
	return template.Must(template.New("layout.html").Funcs(template.FuncMap{
		"contains": strings.Contains,
	}).ParseFiles(
		templatePath+"/layout.html",
		templatePath+"/"+name,
	))
}

type LayoutTemplateData struct {
	Common NestedCommonTplData
	Login  NestedLoginTplData
	Data   any
}

func execTemplate(w http.ResponseWriter, t *template.Template, commonData NestedCommonTplData, loginData NestedLoginTplData, config ConfigFile, data any) error {
	commonData.WebsiteURL = config.WebAddress
	commonData.WebsiteName = config.Org
	return t.Execute(w, LayoutTemplateData{
		Common: commonData,
		Login:  loginData,
		Data:   data,
	})
}
