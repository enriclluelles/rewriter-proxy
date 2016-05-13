package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
)

type Config struct {
	SiteConfigs []*SiteConfig `json:"sites"`
	ListenURL   string        `json:"listenURL"`
}

type SiteConfig struct {
	Replacements     []*ReplacementRule `json:"replacements"`
	HostHeader       string             `json:"hostHeader"`
	VirtualHost      string             `json:"virtualHost"`
	Endpoint         string             `json:"endpoint"`
	RewriteLocations bool               `json:"rewriteLocations"`
	EndpointURL      *url.URL
}

type ReplacementRule struct {
	SourceString string `json:"source"`
	Target       string `json:"target"`
	Source       *regexp.Regexp
}

func parseConfig() Config {
	var config Config

	var filename string

	if len(os.Args) < 2 {
		//If filename is not provided use ./config.json
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Fatal(err)
		}

		filename = path.Join(dir, "config.json")
	} else {
		filename = os.Args[1]
	}

	configString, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Coulnd't open file %s: %#v", filename, err)
	}

	err = json.Unmarshal(configString, &config)
	if err != nil {
		log.Fatal("Couldn't parse file %s: %#v", filename, err)
	}

	for _, s := range config.SiteConfigs {

		s.EndpointURL, err = url.Parse(s.Endpoint)

		for _, r := range s.Replacements {
			r.Source = regexp.MustCompile(r.SourceString)
		}
	}

	return config
}
