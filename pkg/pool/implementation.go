package pool

import (
	"time"

	"github.com/Pushwoosh/go-connection-pool/pkg/connection"
	"github.com/Pushwoosh/go-connection-pool/pkg/message"
)

type Config struct {
	MaxConnections int
	CheckInterval  time.Duration
	Dialer         connection.Dialer
}

type Pool struct {
	config       Config
	ticker       *time.Ticker
	connections  *connection.Connections
	internalInCh chan message.Message
}

func NewPool(c Config) *Pool {
	p := new(Pool)
	p.config = c
	p.internalInCh = make(chan message.Message) // not buffered
	p.ticker = time.NewTicker(p.config.CheckInterval)
	p.connections = connection.NewConnections(p.config.MaxConnections)
	return p
}

func (p *Pool) makeConnections(out chan message.Message) {
	for p.connections.Len() < p.config.MaxConnections {
		conn, err := p.config.Dialer()
		// TODO: limit retry here
		if err != nil {
			continue
		}
		p.connections.Add(conn)
		go func() {
			conn.Serve(p.internalInCh, out)
		}()
	}
}

// blocked call, closing `in`-chan mean stop Serve
func (p *Pool) Serve(in chan message.Message, out chan message.Message) error {
	go p.makeConnections(out)

	go func() {
		for range p.ticker.C {
			// iterate over all connections and remove all not live
			_ = p.connections.Clean()

			// restore connection's num
			p.makeConnections(out)
		}
	}()

	for m := range in {
		p.internalInCh <- m
	}

	p.ticker.Stop()
	// signal for connections to stop
	close(p.internalInCh)
	// wait until all connections complete work
	for p.connections.Len() != 0 {
		if err := p.connections.Clean(); err != nil {
			// internal err which must not happens ;)
			return err
		}
	}
	return nil
}
