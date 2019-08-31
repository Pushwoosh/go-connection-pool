package pool

import (
	"fmt"
	"testing"
	"time"

	"github.com/Pushwoosh/go-connection-pool/pkg/connection"
	"github.com/Pushwoosh/go-connection-pool/pkg/message"
)

type msg struct {
}

type conn struct {
	State int
	Id    int
}

func (c *conn) String() string {
	return fmt.Sprintf("<%d,%t>", c.Id, c.Live())
}

func (c *conn) Live() bool {
	c.State += 1
	return c.State < 6
}

func (c *conn) Serve(in chan message.Message, out chan message.Message) {
	for m := range in {
		out <- m
	}
	c.State = 6
}

func Test_Connections_Serve(t *testing.T) {
	items := 10000

	p := NewPool(Config{
		MaxConnections: 123,
		CheckInterval:  10 * time.Millisecond,
		Dialer: func() (connection.Connection, error) {
			return &conn{}, nil
		},
	})

	inChan := make(chan message.Message)
	outChan := make(chan message.Message)

	go func() {
		for count := 0; count < items; count++ {
			inChan <- msg{}
		}
		close(inChan)
	}()

	go func() {
		_ = p.Serve(inChan, outChan)
		close(outChan)
	}()

	counter := 0
	for range outChan {
		counter += 1
	}

	if counter != items {
		t.Fatalf("Some items lost %d != %d", counter, items)
	}
}
