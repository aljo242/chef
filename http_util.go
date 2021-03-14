package http_util

import (
	"fmt"
	"log"
	"net/http"
)

func Print() error {
	fmt.Println("vim-go")
	return nil
}

// CheckHTTP2Support is a simple test to see if HTTP2 is supported by checking if http.Pusher is in the responsewriter
func CheckHTTP2Support(w http.ResponseWriter) bool {
	_, ok := w.(http.Pusher)
	if ok {
		log.Printf("HTTP/2 Supported!\n")
	} else {
		log.Printf("HTTP/2 NOT Supported!\n")
	}

	return ok
}

// RedirectHTTPS can redirect all http traffic to corresponding https addresses
func RedirectHTTPS(httpsHost string, debugEnable bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if debugEnable {
			log.Printf("%v\n", httpsHost+r.RequestURI)
		}
		http.Redirect(w, r, httpsHost+r.RequestURI, http.StatusMovedPermanently)
	}
}
