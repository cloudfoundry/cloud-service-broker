// Package freeport identifies a random unused port
package freeport

import (
	"fmt"
	"net"
)

func Port() (int, error) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, fmt.Errorf("unable to open a listener port: %w", err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func Must() int {
	port, err := Port()
	if err != nil {
		panic(err)
	}
	return port
}
