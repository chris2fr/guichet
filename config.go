/*
config handles reading the config.json file at the root and processing the settings
*/
package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
)

var configFlag = flag.String("config", "./config.json", "Configuration file path")

var config *ConfigFile

func readConfig() ConfigFile {
	// Default configuration values for certain fields
	config_file := ConfigFile{
		HttpBindAddr:   ":9991",
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
