package http_util

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

const (
	sampleConfigFile    = "./sample/sample_config.json"
	sampleHTML          = "./sample/test.html"
	incorrectConfigFile = "incorrect.wrong"
)

var (
	client *http.Client
)

func pushAttemptHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
	err := PushFiles(w, sampleHTML)
	if err != nil {
		log.Error().Err(err).Msg("UNABLE TO PUSH")
	}

	w.WriteHeader(http.StatusOK)
}

func validHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func invalidHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
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
	r.HandleFunc("/valid", validHandler)
	r.HandleFunc("/invalid", invalidHandler)
	r.HandleFunc("/pushAttempt", pushAttemptHandler)

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

func TestConfig(t *testing.T) {
	// provide nonexistent file to get incorrect file error
	_, err := LoadConfig(incorrectConfigFile)
	if err != os.ErrNotExist {
		t.Errorf("error loading non-existent file into config : %v", err)
	}

	_, err = LoadConfig(sampleHTML)
	if err != ErrConfigNotJSON {
		t.Errorf("error loading non-json file into config : %v", err)
	}

}

func TestValidGetRequest(t *testing.T) {
	wantStatus := "200 OK"

	r, err := client.Get("http://localhost/valid")
	if err != nil {
		fmt.Println(r)
		t.Errorf("Error with valid get request to server : %v", err)
	}
	defer r.Body.Close()

	assert.Equal(t, wantStatus, r.Status)
}

func TestInvalidGetRequest(t *testing.T) {
	wantStatus := "404 Not Found"

	r, err := client.Get("http://localhost/invalid")
	if err != nil {
		fmt.Println(r)
		t.Errorf("Error with invalid get request to server : %v", err)
	}
	defer r.Body.Close()

	assert.Equal(t, wantStatus, r.Status)
}

func TestPushAttemptRequest(t *testing.T) {
	wantStatus := "200 OK"

	r, err := client.Get("http://localhost/pushAttempt")
	if err != nil {
		fmt.Println(r)
		t.Errorf("Error with invalid get request to server : %v", err)
	}
	defer r.Body.Close()

	assert.Equal(t, wantStatus, r.Status)

}
