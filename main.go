package main

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

func newServer(file, port string) *http.Server {
	CACertPath = file
	router := http.NewServeMux()
	router.HandleFunc("/", serveFile)
	router.Handle("/metrics", promhttp.Handler())

	return &http.Server{
		Addr:    			":" + port,
		Handler: router,
	}
}

func serveFile(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadFile(CACertPath)
	if err != nil {
		log.Fatal(err)
	}
	http.ServeContent(w, r, "ca.crt", time.Now(), bytes.NewReader(data))
}

func startMetricsCollector() {
	collectMetrics()

	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			collectMetrics()
		}
	}
}

func collectMetrics() {
	expiryTimestamp, err := getCertExpiryTimestamp(CACertPath)
	if err != nil {
		log.Printf("Error getting certificate expiry: %v\n", err)
		return
	}
	CACertExpiryMetric.Set(float64(expiryTimestamp.Unix()))
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

func listenForShutdown(ctx context.Context, server *http.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Printf("Shutting down")

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	if err != nil {
		log.Println(err, "Failed to shutdown http server")
	}
}

func main() {
	flag.Parse()
	prometheus.MustRegister(CACertExpiryMetric)

	server := newServer(*file, *port)
	go listenForShutdown(context.Background(), server)
	go startMetricsCollector()
	log.Printf("Serving %s on HTTP port: %s\n", *file, *port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}
