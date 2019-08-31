package connection

import (
	"fmt"
	"testing"

	"github.com/Pushwoosh/go-connection-pool/pkg/message"
)

type conn struct {
	State bool
	Id    int
}

func (c *conn) String() string {
	return fmt.Sprintf("<%d,%t>", c.Id, c.Live())
}

func (c *conn) Live() bool {
	return c.State
}

func (c *conn) Serve(in chan message.Message, out chan message.Message) {
	for m := range in {
		out <- m
	}
	c.State = false
}

const defaultSize = 5

func Test_Connections_New(t *testing.T) {
	conns := NewConnections(defaultSize)

	if conns.Len() != 0 {
		t.Fatalf("Len must be 0, capacity != 0")
	}

	if conns.cap() != defaultSize {
		t.Fatalf("Capacity != %d", defaultSize)
	}
}

func Test_Connections_Add(t *testing.T) {
	numForAdd := 10

	conns := NewConnections(defaultSize)

	for index := 1; index <= numForAdd; index += 1 {
		conns.Add(&conn{Id: index})

		if conns.Len() != index {
			t.Fatalf("Len must be %d, but it's %d; %s", index, conns.Len(), conns)
		}
	}
}

func Test_Connections_Remove(t *testing.T) {
	numForAdd := 10

	conns := NewConnections(defaultSize)

	for index := 0; index < numForAdd; index += 1 {
		conns.Add(&conn{Id: index})
	}

	_, _ = conns.Pop(0)
	if conns.String() != "[Len=9, Cap=10; <1,false>,<2,false>,<3,false>,<4,false>,<5,false>,<6,false>,<7,false>,<8,false>,<9,false>]." {
		t.Fatalf("Remove incorrect elem, index=%d, %s", 0, conns)
	}

	_, err := conns.Pop(10)
	if err == nil {
		t.Fatalf("Remove unexisted elem, index=%d, %s", 10, conns)
	}

	_, _ = conns.Pop(7)
	if conns.String() != "[Len=8, Cap=10; <1,false>,<2,false>,<3,false>,<4,false>,<5,false>,<6,false>,<7,false>,<9,false>]." {
		t.Fatalf("Remove incorrect elem, index=%d, %s", 10, conns)
	}
}

func Test_Connections_Clean(t *testing.T) {
	numForAdd := 11

	conns := NewConnections(defaultSize)

	var state bool
	for index := 0; index < numForAdd; index += 1 {
		if index%2 == 0 {
			state = true
		} else {
			state = false
		}
		conns.Add(&conn{Id: index, State: state})
	}

	if err := conns.Clean(); err != nil {
		t.Fatalf("%s: %s", err.Error(), conns)
	}

	if conns.Len() != 6 {
		t.Fatalf("Clean work incorrect, %s", conns)
	}
}
