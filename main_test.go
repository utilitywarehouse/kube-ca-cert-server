package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func readFromTestServer() ([]byte, error) {
	resp, err := http.Get("http://localhost:8080/")
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// TestListenAndServeFileUpdate checks that updating a file's content will make
// the server to actually refresh and serve the updated
func TestListenAndServeFileUpdate(t *testing.T) {
	content := []byte("test file's content")
	tmpfile, err := ioutil.TempFile("", "test-file")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write(content); err != nil {
		log.Fatal(err)
	}
	go listenAndServe(tmpfile.Name(), "8080")
	time.Sleep(1) // Sleep to allow some time for the server to come up
	// Get from server to verify we are serving the content
	resp, err := readFromTestServer()
	if err != nil {
		log.Fatal(err)
	}
	// To string for human friendly output on error
	assert.Equal(t, content, resp)

	tmpfile.Truncate(0)
	tmpfile.Seek(0, 0)
	content = []byte("updated test file's content")
	if _, err := tmpfile.Write(content); err != nil {
		log.Fatal(err)
	}
	// Get from server to verify that the server responds with the latest
	// updated content
	resp, err = readFromTestServer()
	if err != nil {
		log.Fatal(err)
	}
	// To string for human friendly output on error
	assert.Equal(t, string(content), string(resp))

	// Close before leaving
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}
}
