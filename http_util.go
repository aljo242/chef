package http_util

import (
	"fmt"
	"net/http"
	"log"
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
