package connection

import "github.com/Pushwoosh/go-connection-pool/pkg/message"

// all methods must be blocking
type Connection interface {
	// make sure that after closing Serve's `in`-chan
	// and complete processing all messages it method
	// will return false
	Live() bool
	Serve(chan message.Message, chan message.Message)
}

// closure with ready connection configuration
// must return Connection with `Live` state true
type Dialer func() (Connection, error)
