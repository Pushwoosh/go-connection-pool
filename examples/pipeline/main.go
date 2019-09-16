package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/Pushwoosh/go-connection-pool/pkg/connection"
	"github.com/Pushwoosh/go-connection-pool/pkg/message"
	"github.com/Pushwoosh/go-connection-pool/pkg/pool"
)

const defaultTime = time.Second * 5

type msg struct {
	Destination   string
	DataTemplate  string
	TemplatedData string
	Status        string
}

type NameGenerator func() string

func ConcreteNameGenerator(namesList []string) NameGenerator {
	len := len(namesList) - 1
	return func() string {
		// nolint: gosec
		return namesList[rand.Int()%len]
	}
}

type MsgGenerator func() msg

func ConcreteMgsGenerator(nameGenerator NameGenerator) MsgGenerator {
	return func() msg {
		return msg{
			// nolint: gosec
			DataTemplate: "Hi, %s",
			Destination:  fmt.Sprintf("https://duckduckgo.com/?q=%s&t=h_&ia=web", nameGenerator()),
		}
	}
}

type FakeRabbitMQDataProvider struct {
	FakeMsgNum   int
	MsgGenerator MsgGenerator
}

func (r *FakeRabbitMQDataProvider) Serve(in chan message.Message, out chan message.Message) {
	<-in
	for count := 0; count < r.FakeMsgNum; count++ {
		out <- r.MsgGenerator()
	}
	close(out)
}

type AdditionalDataProvider struct {
	NameGenerator NameGenerator
}

func (d *AdditionalDataProvider) Serve(in chan message.Message, out chan message.Message) {
	for untypedM := range in {
		m, ok := untypedM.(msg)
		if !ok {
			continue
		}
		m.TemplatedData = fmt.Sprintf(m.DataTemplate, d.NameGenerator())
		out <- m
	}
	close(out)
}

type conn struct {
	realConn http.Client
	ID       int
	State    bool
}

func (c *conn) Live() bool {
	return c.State
}

func (c *conn) Serve(in chan message.Message, out chan message.Message) {
	for untypedM := range in {
		m, ok := untypedM.(msg)
		if !ok {
			continue
		}
		if resp, err := c.realConn.Get(m.Destination); err != nil || resp.StatusCode != 200 {
			m.Status = "FAIL"
		} else {
			m.Status = "OK"
		}
		out <- m
	}
	c.State = false
}

func poolConnectionDialer() connection.Dialer {
	counter := 0
	return func() (connection.Connection, error) {
		counter++
		c := &conn{
			ID:    counter,
			State: true,
			realConn: http.Client{
				Timeout: defaultTime,
				Transport: &http.Transport{
					Dial: (&net.Dialer{
						Timeout: defaultTime,
					}).Dial,
					TLSHandshakeTimeout: defaultTime,
				},
			},
		}
		log.Printf("New connection: %d\n", counter)
		return c, nil
	}
}

type StatsReporter struct{}

func (s *StatsReporter) Serve(in chan message.Message, out chan message.Message) {
	for untypedM := range in {
		m, ok := untypedM.(msg)
		if !ok {
			continue
		}
		log.Printf("%s: %s\n", m.TemplatedData, m.Status)
	}
	out <- msg{}
}

func main() {
	nameGenerator := ConcreteNameGenerator([]string{
		"Alice",
		"Bob",
	})
	msgGenerator := ConcreteMgsGenerator(nameGenerator)

	rabbit := &FakeRabbitMQDataProvider{
		FakeMsgNum:   5,
		MsgGenerator: msgGenerator,
	}
	additionalDataProvider := &AdditionalDataProvider{
		NameGenerator: nameGenerator,
	}
	pool := pool.NewPool(pool.Config{
		MaxConnections: 5,
		CheckInterval:  defaultTime,
		Dialer:         poolConnectionDialer(),
	})
	stats := &StatsReporter{}

	inRabbitCh := make(chan message.Message)
	outRabbitCh := make(chan message.Message)
	outAdditionalDataProviderCh := make(chan message.Message)
	outPoolCh := make(chan message.Message)
	statsOutCh := make(chan message.Message)

	go rabbit.Serve(inRabbitCh, outRabbitCh)
	go additionalDataProvider.Serve(outRabbitCh, outAdditionalDataProviderCh)
	go func() { _ = pool.Serve(outAdditionalDataProviderCh, outPoolCh); close(outPoolCh) }()
	go stats.Serve(outPoolCh, statsOutCh)

	inRabbitCh <- msg{}
	<-statsOutCh

}
