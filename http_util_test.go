package http_util

import (
	"testing"

	"github.com/gorilla/mux"
)

const (
	sampleConfigFile = "./sample/sample_config.json"
)

func TestBasic(t *testing.T) {
	runningChan := make(chan struct{})

	cfg, err := LoadConfig(sampleConfigFile)
	if err != nil {
		t.Fatal(err)
	}

	r := mux.NewRouter()
	// attach basic handler
	r.HandleFunc("/", basicTestHandler)

	srv := NewServer(cfg, r)
	go func(ch chan struct{}) {
		srv.Run(ch)
	}(runningChan)

	// wait until running message
	<-runningChan

	err = srv.Quit()
	if err != nil {
		t.Fatal(err)
	}
}
