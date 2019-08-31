package connection

import (
	"fmt"
	"strings"
	"sync"
)

type Connections struct {
	sync.RWMutex
	slice []Connection
}

func (c *Connections) len() int {
	return len(c.slice)
}

func (c *Connections) cap() int {
	return cap(c.slice)
}

func (c *Connections) pop(index int) (Connection, error) {
	if index >= c.len() {
		return nil, fmt.Errorf("%d out of bounds", index)
	}
	elem := c.slice[index]
	c.slice = append(c.slice[:index], c.slice[index+1:]...)
	return elem, nil
}

func (c *Connections) Add(conn Connection) {
	c.Lock()
	defer c.Unlock()
	c.slice = append(c.slice, conn)
}

func (c *Connections) Len() int {
	c.RLock()
	defer c.RUnlock()
	return c.len()
}

func (c *Connections) Pop(index int) (Connection, error) {
	c.Lock()
	defer c.Unlock()
	return c.pop(index)
}

func (c *Connections) String() string {
	c.RLock()
	defer c.RUnlock()
	r := make([]string, 0, c.len())
	for _, conn := range c.slice {
		if connStringer, ok := conn.(fmt.Stringer); ok {
			r = append(r, connStringer.String())
		} else {
			r = append(r, fmt.Sprintf("%#v", conn))
		}
	}
	return fmt.Sprintf("[Len=%d, Cap=%d; %s].", c.len(), c.cap(), strings.Join(r, ","))
}

func (c *Connections) Clean() (err error) {
	c.Lock()
	defer c.Unlock()
RereadSlice:
	for index, conn := range c.slice {
		if conn.Live() {
			continue
		}
		if _, err = c.pop(index); err != nil {
			return
		}
		// if we successfully remove item this slice is inconsistent
		goto RereadSlice
	}
	return
}

// it's just data-structure without logic
func NewConnections(size int) *Connections {
	c := new(Connections)
	c.slice = make([]Connection, 0, size)
	return c
}
