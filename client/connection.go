package client

import (
	"net"
	"crypto/tls"
)

// Connect connects the server via specified network
func (c *Client) Connect(network, address string) error {
	var conn net.Conn
	var err error

	switch network {
	case "http":
		conn, err =
	}
}

func newDirectHTTPConn(c *Client, network, address string) (net.Conn, error) {
	path := c.option
	if path == "" {
		return nil,error("")
	}

	var conn net.Conn
	var tlsConn *tls.Conn
	var err error

	if c != nil && c.option
}
