package main

import (
	"flag"
	"fmt"
	"html"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/Pushwoosh/go-connection-pool/pkg/connection"
	"github.com/Pushwoosh/go-connection-pool/pkg/message"
	"github.com/Pushwoosh/go-connection-pool/pkg/pool"
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

type conn struct {
	realConn http.Client
	Id       int
	State    bool
}

func (c *conn) String() string {
	return fmt.Sprintf("<%d,%t>", c.Id, c.Live())
}

func (c *conn) Live() bool {
	return c.State
}

func (c *conn) Serve(in chan message.Message, out chan message.Message) {
	for m := range in {
		mTyped, ok := m.(msg)
		if !ok {
			continue
		}
		log.Printf("[%s] sending: %#v\n", c, m)
		var err error
		if mTyped.resp, err = c.realConn.Get(mTyped.url); err != nil || mTyped.resp.StatusCode != 200 {
			mTyped.status = "FAIL"
		} else {
			mTyped.status = "OK"
		}
		out <- mTyped
	}
	log.Printf("[%s] close me!\n", c)
	c.State = false
}

type msg struct {
	url    string
	status string
	resp   *http.Response
}

var (
	reqNum          = flag.Int("req-num", 1000, "Num of requests")
	connNum         = flag.Int("conn-num", 100, "Num of parallel connections")
	url             = flag.String("url", "http://localhost:8080/hello", "Url to requests")
	timeoutDuration = time.Second * 5
)

func main() {
	flag.Parse()

	counter := 0
	cfg := pool.Config{
		MaxConnections: *connNum,
		CheckInterval:  timeoutDuration,
		Dialer: func() (connection.Connection, error) {
			counter += 1
			c := &conn{
				Id:    counter,
				State: true,
				realConn: http.Client{
					Timeout: timeoutDuration,
					Transport: &http.Transport{
						Dial: (&net.Dialer{
							Timeout: timeoutDuration,
						}).Dial,
						TLSHandshakeTimeout: timeoutDuration,
					},
				},
			}
			log.Printf("New connection: %s\n", c)
			return c, nil
		},
	}
	p := pool.NewPool(cfg)

	log.Printf("Started with reqNum=%d, url=%s cfg=%#v", *reqNum, *url, cfg)
	defer log.Println("Completed")

	inChan := make(chan message.Message)
	outChan := make(chan message.Message)

	go func() {
		for count := 0; count < *reqNum; count++ {
			inChan <- msg{url: *url}
		}
		close(inChan)
	}()

	go func() {
		count := 0
		for m := range outChan {
			count += 1
			mTyped, ok := m.(msg)
			if !ok {
				continue
			}
			log.Printf("%d) Receive: %#v\n", count, mTyped.resp)
		}
	}()

	if *url == "http://localhost:8080/hello" {
		startWebServer()
	}

	log.Println("Start serving")
	if err := p.Serve(inChan, outChan); err != nil {
		log.Printf("%#v\n", err)
	}
	close(outChan)
}
