package http_util

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

const (
	sampleConfigFile = "./sample/sample_config.json"
	sampleHTML       = "./sample/test.html"
)

var (
	client *http.Client
)

func basicTestHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
	err := PushFiles(w, sampleHTML)
	if err != nil {
		log.Error().Err(err).Msg("UNABLE TO PUSH")
	}
}

func TestMain(m *testing.M) {
	runningChan := make(chan struct{})

	cfg, err := LoadConfig(sampleConfigFile)
	if err != nil {
		os.Exit(-1)
	}
	cfg.Print()

	r := mux.NewRouter()
	// attach basic handler
	r.HandleFunc("/valid", basicTestHandler)

	srv := NewServer(cfg, r)
	go func(ch chan struct{}) {
		srv.Run(ch)
	}(runningChan)

	// wait until running message
	<-runningChan
	fmt.Printf("server is running on: %v\n", cfg.Host+":"+cfg.Port)
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client = &http.Client{Transport: tr}
	time.Sleep(3 * time.Second)

	exitCode := m.Run()

	err = srv.Quit()
	if err != nil {
		os.Exit(-1)
	}

	if srv.isRunning {
		os.Exit(-1)
	}
	os.Exit(exitCode)
}

func TestValidGetRequest(t *testing.T) {
	r, err := client.Get("http://localhost/valid")
	if err != nil {
		fmt.Println(r)
		t.Errorf("Error with valid get request to server : %v", err)
	}
	defer r.Body.Close()
}

func TestInvalidGetRequest(t *testing.T) {
	r, err := client.Get("http://localhost/")
	if err != nil {
		fmt.Println(r)
		t.Errorf("Error with invalid get request to server : %v", err)
	}
	defer r.Body.Close()
}
