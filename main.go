package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path"
	"sync"
)

type config struct {
	ListenAddr string `json:"listenAddress"`
	ListenPort int    `json:"listenPort"`
	DataRoot   string `json:"dataroot"`

	m sync.Mutex // to protect the shelling out from having races
}

var (
	// listenAddr = flag.String("listenAddr", "[::1]", "IPv4 or IPv6 address to listen on")
	// listenPort = flag.Int("listenPort", 8080, "Numeric port to listen on")
	configFile = flag.String("configFile", "", "JSON configuration file to read config from")

	debug = log.New(os.Stderr, "debug ", log.LstdFlags|log.Lshortfile|log.LUTC|log.Lmicroseconds)

	addr string
)

func init() {
	flag.Parse()
	if *configFile == "" {
		debug.Fatal("`-configFile' is a required option[sic!]")
	}
}

func main() {
	debug.Printf("Initializing...")
	c, err := loadAndInitialize(*configFile)
	if err != nil {
		debug.Fatalf("loadAndInitialize: %q", err)
	}

	addr = fmt.Sprintf("%s:%d", c.ListenAddr, c.ListenPort)
	debug.Printf("Done initializing!")

	debug.Printf("Listening on http://%s", addr)

	http.HandleFunc("/api/upload", c.upload)
	http.ListenAndServe(addr, nil)
}

func loadAndInitialize(f string) (*config, error) {
	rdr, err := os.Open(f)
	if err != nil {
		debug.Fatalf("os.Open: %q", err)
	}
	b, err := ioutil.ReadAll(rdr)
	if err != nil {
		debug.Fatalf("ioutil.ReadAll: %q", err)
	}

	var c config
	err = json.Unmarshal(b, &c)
	if err != nil {
		debug.Fatalf("json.Unmarshal: %q", err)
	}

	err = os.MkdirAll(c.DataRoot, 0755)
	if err != nil {
		debug.Fatalf("os.MkdirAll: %q", err)
	}
	return &c, err
}

func (c config) upload(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received a %q reqeust", r.Method)
	switch {
	case r.Method == "POST":
		r.ParseMultipartForm(int64(math.Pow(2, 20) * 100))
		file, header, err := r.FormFile("upload")
		if err != nil {
			log.Printf("err(http.Request.FormFile): %q", err)
			return
		}
		defer file.Close()
		log.Printf("header: %q", header.Header)
		if err != nil {
			log.Printf("err(ioutil.TempFile): %q", err)
			return
		}
		newName := path.Join(c.DataRoot, header.Filename)
		log.Printf("newName: %q", newName)
		wr, err := os.OpenFile(newName, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		io.Copy(wr, file)
		if err != nil {
			log.Printf("err(os.Rename): %q", err)
		}
		wr.Close()
		err = c.writePackages()
		if err != nil {
			log.Printf("err(c.writePackages): %q", err)
		}
		fmt.Fprintln(w, "OK i have written the file!")

	default:
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

}

func (c config) writePackages() error {
	c.m.Lock()
	defer c.m.Unlock()
	cmd := exec.Command(
		"R",
		"-e",
		fmt.Sprintf(`'tools::write_PACKAGES(dir="%s", subdirs = TRUE, latestOnly = FALSE, addFiles = TRUE, verbose = TRUE)'`, c.DataRoot),
	)
	debug.Printf("cmd.Args: %q", cmd.Args)
	err := cmd.Run()
	if err != nil {
		debug.Printf("err(cmd.Output): %q", err)
	}
	return err
}
