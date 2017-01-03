package main

import (
	"fmt"
	"net/http"
)

func startManagerServer() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", manager)

	server := &http.Server{
		Addr:    ":7000",
		Handler: mux,
	}

	server.ListenAndServe()

}

func manager(w http.ResponseWriter, r *http.Request) {

	msg := <-RequestChan

	fmt.Fprintf(w, "This is manager: %s", msg)
}
