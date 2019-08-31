package main

import (
	"flag"
	"fmt"
	"html"
	"log"
	"net"
	"net/http"
	"time"
)

func startWebServer() {
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Millisecond * 50)
		_, _ = fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	go func() {
		err := http.ListenAndServe("localhost:8080", nil)
		if err != nil {
			log.Printf("%#v", err)
		}
	}()
}

var (
	reqNum          = flag.Int("req-num", 1000, "Num of requests")
	url             = flag.String("url", "http://localhost:8080/hello", "Url to requests")
	timeoutDuration = time.Second * 5
)

func main() {
	flag.Parse()

	log.Printf("Started with reqNum=%d, url=%s", *reqNum, *url)
	defer log.Println("Completed")

	conn := http.Client{
		Timeout: timeoutDuration,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: timeoutDuration,
			}).Dial,
			TLSHandshakeTimeout: timeoutDuration,
		},
	}

	if *url == "http://localhost:8080/hello" {
		startWebServer()
	}

	for counter := 0; counter < *reqNum; counter += 1 {
		resp, err := conn.Get(*url)
		log.Printf("%#v, %#v", resp, err)
	}
}
