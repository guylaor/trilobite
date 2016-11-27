package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
)

func main() {
	fmt.Println("Trilobite Debugging proxy")

	cmdLocalPort := flag.String("port", "8888", "Local port for incoming connections")
	flag.Parse()

	localPort := fmt.Sprintf(":%s", *cmdLocalPort)
	// listen for connections forver
	ln, err := net.Listen("tcp", localPort)
	if err != nil {
		log.Fatalf("Failed to list on: %s", err)
	}
	fmt.Printf("Listening on localhost port: %s \n", *cmdLocalPort)
	// endless loop to accept connections
	for {
		conn, err := ln.Accept()
		if err == nil {
			go handleConnection(conn)
		} else {
			log.Fatalf("failed to accept connection %s", err)
		}
	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		req, err := http.ReadRequest(reader)
		if err != nil {
			if err == io.EOF {
				log.Printf("Failed to read request: %s", err)
			}
			return
		}

		log.Printf("request: %s %s %s %s", req.URL.Path, req.Host, req.Method, req.URL.String())

		backendUrl := fmt.Sprintf("%s:%s", req.Host, "80")
		log.Printf("backendurl %S", backendUrl)
		// sending the request to backend
		if be, err := net.Dial("tcp", backendUrl); err == nil {
			be_reader := bufio.NewReader(be)
			if err := req.Write(be); err == nil {
				// read the response from the backend
				if resp, err := http.ReadResponse(be_reader, req); err == nil {
					resp.Close = true
					if err := resp.Write(conn); err == nil {
						log.Printf("%s: %d", req.URL.Path, resp.StatusCode)
					}
					conn.Close()
				}
			}
		}

	}
}
