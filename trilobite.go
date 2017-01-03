package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type RequestMsg struct {
	url          string
	ResponseBody string
}

var RequestChan chan RequestMsg

type Filters struct {
	Requests []struct {
		Pattern string `json:"pattern"`
	} `json:"requests"`
}

var loadedFilters Filters

func main() {

	cmdLocalPort := flag.String("port", "8888", "Local port for incoming connections")
	flag.Parse()

	fmt.Printf("\n\nTrilobite Debugging Proxy - using port %s\n\n", *cmdLocalPort)

	localPort := fmt.Sprintf(":%s", *cmdLocalPort)

	loadedFilters = loadFilters()

	RequestChan = make(chan RequestMsg)

	go startManagerServer()

	//go func() {
	http.HandleFunc("/", HandleConnections)
	log.Fatal(http.ListenAndServe(localPort, nil))
	//}()

	// var input string
	// fmt.Scanln(&input)
	// fmt.Println("done")
}

func detectTextContentType(url string, contentType *string) {
	if strings.Contains(*contentType, "text/plain") {
		if strings.Contains(url, "css") {
			*contentType = "text/css"
		}
		if strings.Contains(url, ".js") {
			*contentType = "application/javascript"
		}
	}
}

func loadFilters() Filters {

	content, err := ioutil.ReadFile("filters.json")
	if err != nil {
		log.Panic(err)
	}

	var filters Filters
	if err := json.Unmarshal(content, &filters); err != nil {
		log.Panic(err)
	}

	log.Printf("Loaded %s filters \n", len(filters.Requests))

	return filters
}

func HandleConnections(w http.ResponseWriter, req *http.Request) {

	//log.Printf("Request: %s \n", req.URL.String())

	// making a copy of request headers
	headers := map[string][]string{}
	headers = req.Header

	// building the request again
	newreq, err := http.NewRequest(req.Method, req.URL.String(), req.Body)
	if err != nil {
		log.Print(err)
	}

	var client = &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Do(newreq)
	if err != nil {
		log.Print(err)
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print(err)
	}

	//cotcopy := content
	for _, filter := range loadedFilters.Requests {
		match, _ := regexp.MatchString(filter.Pattern, req.URL.String())
		if match {
			msg := RequestMsg{req.URL.String(), fmt.Sprintf("%s", content)}
			log.Printf("msg:%s \n", msg.url)
			RequestChan <- msg
		}
	}

	// get the "right" content type, and then get it again for css and js
	contentType := http.DetectContentType(content)
	detectTextContentType(req.URL.String(), &contentType)

	//log.Printf("Response: %s \n", resp)
	//log.Printf("Copy size: %d, resp size:%d \n", len(cotcopy), len(content))

	// adding headers to the request
	w.Header().Set("Content-Type", contentType)
	for k, v := range headers {
		w.Header().Set(k, strings.Join(v, " "))
	}

	// writing back the Response
	fmt.Fprintf(w, "%s", content)

}
