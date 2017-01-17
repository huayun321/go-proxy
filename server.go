// server.go
package main

import (
	"fmt"
	"net"
)

func main() {
	// 1. Listen for connections forever.
	if ln, err := net.Listen("tcp", ":8080"); err == nil {

		//2. Accept connections.
		if conn, err := ln.Accept(); err == nil {
			// ...more code...
			fmt.Println("accept connection")
		}
	}
}
