package main

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const defaultCACertPath = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"

var CACertPath string
var (
	port = flag.String("p", "8080", "port to serve on")
	file = flag.String("f", defaultCACertPath, "the static file to host")
)

var (
	CACertExpiryMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "ca_cert_expiry_timestamp",
		Help: "Timestamp of CA certificate expiration in seconds since UNIX epoch",
	})
)

func serveFile(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadFile(CACertPath)
	if err != nil {
		log.Fatal(err)
	}
	http.ServeContent(w, r, "ca.crt", time.Now(), bytes.NewReader(data))
}

func listenAndServe(file, port string) {
	CACertPath = file
	http.HandleFunc("/", serveFile)
	http.Handle("/metrics", promhttp.Handler())

	log.Printf("Serving %s on HTTP port: %s\n", file, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func collectMetrics() {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			expiryTimestamp, err := getCertExpiryTimestamp(CACertPath)
			if err != nil {
				log.Printf("Error getting certificate expiry: %v\n", err)
				continue
			}
			CACertExpiryMetric.Set(float64(expiryTimestamp.Unix()))
			log.Printf("CA certificate expiry: %s\n", expiryTimestamp)
		}
	}
}

func getCertExpiryTimestamp(certFile string) (time.Time, error) {
	data, err := ioutil.ReadFile(certFile)
	if err != nil {
		return time.Time{}, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return time.Time{}, fmt.Errorf("failed to parse PEM block")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, err
	}

	return cert.NotAfter, nil
}

func main() {
	flag.Parse()
	prometheus.MustRegister(CACertExpiryMetric)
	go collectMetrics()
	listenAndServe(*file, *port)
}
