package chef

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

const (
	sampleConfigFile    = "./sample/sample_config.json"
	sampleConfigFileTLS = "./sample/sample_config_tls.json"
	sampleHTML          = "./sample/test.html"
	incorrectConfigFile = "incorrect.wrong"
)

var (
	client *http.Client
)

func init() {
	os.Setenv("GODEBUG", "x509ignoreCN=0")
}

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
	if !strings.Contains(os.Getenv("GODEBUG"), "x509ignoreCN=0") {
		fmt.Println("Please set GODEBUG=\"x509ignoreCN=0\" or testing will not work")
		os.Exit(1)
	}

	runningChan := make(chan struct{})

	cfg, err := LoadConfig(sampleConfigFileTLS)
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

	caCert, err := ioutil.ReadFile(cfg.RootCA)
	if err != nil {
		os.Exit(-11)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	fmt.Printf("server is running on: %v\n", cfg.Host+":"+cfg.Port)

	client = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: true,
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}
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

func TestTLSConfig(t *testing.T) {
	// test loading default config with no TLS
	cfg, err := LoadConfig(sampleConfigFile)
	if err != nil {
		t.Error(err)
	}

	// will throw error since no key pair is not present in config
	_, err = getTLSConfig(cfg)
	if err != os.ErrNotExist { // should be returned if no PEM files found in getTLSConfig
		t.Error(err)
	}

	///////////////////////////////////////////////////////////////

	// test loading default config with  TLS
	cfg, err = LoadConfig(sampleConfigFileTLS)
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("Opening %v...\n", cfg.KeyFile)
	f, err := os.Open(cfg.KeyFile)
	if err != nil {
		t.Error(err)
	}
	f.Close()

	fmt.Printf("Opening %v...\n", cfg.CertFile)
	f, err = os.Open(cfg.CertFile)
	if err != nil {
		t.Error(err)
	}
	f.Close()

	fmt.Printf("Loading X509 key pair...\n")
	_, err = tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		t.Error(err)
	}

	_, err = getTLSConfig(cfg)
	if err != nil {
		t.Error(err)
	}

	// test loading default config with TLS but no root CA specified

}

func TestValidGetRequest(t *testing.T) {
	wantStatus := "200 OK"

	fmt.Println("GODEBUG:", os.Getenv("GODEBUG"))
	r, err := client.Get("https://localhost/valid")
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
