package main

import (
	"net"
)

// GetNewAddress gets an available port
func GetNewAddress() (port string, err error) {

	// Create a new server without specifying a port
	// which will result in an open port being chosen
	server, err := net.Listen("tcp4", "127.0.0.1:")

	// If there's an error it likely means no ports
	// are available or something else prevented finding
	// an open port
	if err != nil {
		return "", err
	}

	// Defer the closing of the server so it closes
	defer server.Close()

	// Get the host string in the format "127.0.0.1:4444"
	hostString := server.Addr().String()
	return hostString, nil
}
