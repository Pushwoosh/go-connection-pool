package connection

import "github.com/Pushwoosh/go-connection-pool/pkg/message"

// all methods must be blocking
type Connection interface {
	Live() bool
	Serve(chan message.Message, chan message.Message)
}

// closure with ready connection configuration
type Dialer func() (Connection, error)
