package http_util

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

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

// PushFiles takes an http.ResponseWriter and a variadic amount of file strings
// the function will iterate through each file and performa an HTTP/2 Push
// if HTTP/2 is supported AND if the files are valid. Otherwise, will return error
func PushFiles(w http.ResponseWriter, files ...string) error {
	pusher, ok := w.(http.Pusher)
	if !ok {
		return fmt.Errorf("unable to use http pusher")
	}

	for _, fileName := range files {
		fileName, err := filepath.Abs(filepath.Clean(fileName))
		if err != nil {
			return fmt.Errorf("error getting absolute path: %w", err)
		}
		log.Debug().Str("filename", fileName).Msg("pushing file")
		// TODO add options
		err = pusher.Push(fileName, nil)
		if err != nil {
			return fmt.Errorf("error pushing file %v : %w", fileName, err)
		}
	}

	return nil
}

func basicTestHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
	err := PushFiles(w, "./sample/test.html")
	if err != nil {
		log.Fatal().Err(err).Msg("UNABLE TO PUSH")
	}
}
