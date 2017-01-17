// server.go
package main

import (
	"bufio"
	"log"
	"net"
	"net/http"
)

func main() {
	// 1. Listen for connections forever.
	if ln, err := net.Listen("tcp", ":8080"); err == nil {
		for {

			//2. Accept connections.
			if conn, err := ln.Accept(); err == nil {

				reader := bufio.NewReader(conn)

				//3. Read requests from the client.
				if req, err := http.ReadRequest(reader); err == nil {

					if be, err := net.Dial("tcp", "127.0.0.1:8081"); err == nil {
						be_reader := bufio.NewReader(be)

						if err := req.Write(be); err == nil {

							//6. Read the response from the backend
							if resp, err := http.ReadResponse(be_reader, req); err == nil {

								//7. Send the response to the client.
								resp.Close = true
								if err := resp.Write(conn); err == nil {
									log.Printf("%s: %d", req.URL.Path, resp.StatusCode)
								}
								conn.Close()
								//Repeat back at 2: accept the next connection.

							}
						}
					}

				}

			}
		}

	}
}
