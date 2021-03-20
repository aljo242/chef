package http_util

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

func Print() error {
	fmt.Println("vim-go")
	return nil
}

// CheckHTTP2Support is a simple test to see if HTTP2 is supported by checking if http.Pusher is in the responsewriter
func CheckHTTP2Support(w http.ResponseWriter) bool {
	_, ok := w.(http.Pusher)
	if ok {
		log.Debug().Msg("HTTP/2 Supported!")
	} else {
		log.Debug().Msg("HTTP/2 NOT Supported!")
	}

	return ok
}

// RedirectHTTPS can redirect all http traffic to corresponding https addresses
func RedirectHTTPS(httpsHost string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Str("destination", httpsHost+r.RequestURI).Msg("Redirect HTTPS")
		http.Redirect(w, r, httpsHost+r.RequestURI, http.StatusMovedPermanently)
	}
}

func PushFiles(w http.ResponseWriter, files ...string) error {
	_, ok := w.(http.Pusher)
	if !ok {
		return fmt.Errorf("unable to use http pusher")
	}

	for f := range files {
		fmt.Println(f)
	}

	return nil
}
