package main

import (
	"os"
	"flag"
	"fmt"
	"log"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"encoding/base64"
	"crypto/rand"

	"github.com/gorilla/sessions"
)

type ConfigFile struct {
	HttpBindAddr   string `json:"http_bind_addr"`
	SessionKey     string `json:"session_key"`
	LdapServerAddr string `json:"ldap_server_addr"`
}

var configFlag = flag.String("config", "./config.json", "Configuration file path")

func readConfig() ConfigFile {
	key_bytes := make([]byte, 32)
	n, err := rand.Read(key_bytes)
	if err!= nil || n != 32 {
		log.Fatal(err)
	}

	config_file := ConfigFile{
		HttpBindAddr:   ":9991",
		SessionKey:     base64.StdEncoding.EncodeToString(key_bytes),
		LdapServerAddr: "127.0.0.1:389",
	}

	_, err = os.Stat(*configFlag)
	if os.IsNotExist(err) {
		// Generate default config file
		log.Printf("Generating default config file as %s", *configFlag)

		bytes, err := json.MarshalIndent(&config_file, "", "  ")
		if err != nil {
			log.Fatal(err)
		}

		err = ioutil.WriteFile(*configFlag, bytes, 0644)
		if err != nil {
			log.Fatal(err)
		}

		return config_file
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

var store *sessions.CookieStore = nil

func main() {
	flag.Parse()

	config := readConfig()
	store = sessions.NewCookieStore([]byte(config.SessionKey))

	http.HandleFunc("/", handleHome)

	err := http.ListenAndServe(config.HttpBindAddr, nil)
	if err != nil {
		log.Fatal("Cannot start http server: ", err)
	}
}

// Page handlers ----

func handleHome(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, world!")
}
