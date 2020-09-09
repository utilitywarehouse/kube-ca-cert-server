package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const defaultCACertPath = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"

var (
	port = flag.String("p", "8080", "port to serve on")
	file = flag.String("f", defaultCACertPath, "the static file to host")
)

func serveFile(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadFile(*file)
	if err != nil {
		log.Fatal(err)
	}
	http.ServeContent(w, r, "ca.crt", time.Now(), bytes.NewReader(data))
}

func main() {
	flag.Parse()

	http.HandleFunc("/", serveFile)

	log.Printf("Serving %s on HTTP port: %s\n", *file, *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
