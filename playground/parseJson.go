package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
)

type config struct {
	ListenAddr string `json:"listenAddress"`
	ListenPort int    `json:"listenPort"`
	DataRoot   string `json:"dataroot"`
}

var configFile = flag.String("configFile", "", "Configuration file to parse (JSON formatted)")

func main() {
	flag.Parse()

	rdr, err := os.Open(*configFile)
	if err != nil {
		log.Fatalf("os.Open: %q", err)
	}
	b, err := ioutil.ReadAll(rdr)
	if err != nil {
		log.Fatalf("ioutil.ReadAll: %q", err)
	}

	var c config
	err = json.Unmarshal(b, &c)
	if err != nil {
		log.Fatalf("json.Unmarshal: %q", err)
	}

	log.Printf("Parsed JSON: %#v", c)
}
