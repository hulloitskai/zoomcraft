package minecraft

import (
	"sync"

	"github.com/gorcon/rcon"
)

// A Client is used to communicate with a Minecraft server.
type Client struct {
	conn *rcon.Conn
	mux  sync.Mutex
}

// Execute executes a command on a Minecraft server, and returns the resulting
// output.
//
// It rate-limits and buffers requests to prevent overloading.
func (c *Client) Execute(command string) (out string, err error) {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.conn.Execute(command)
}

// NewClient creates a new Client.
func NewClient(conn *rcon.Conn) *Client {
	return &Client{conn: conn}
}
