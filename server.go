// server.go
package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

//connection pool
type Backend struct {
	net.Conn
	Reader *bufio.Reader
	Writer *bufio.Writer
}

var backendQueue chan *Backend

//statistics
var requestBytes map[string]int64
var requestLock sync.Mutex

func init() {
	requestBytes = make(map[string]int64)
	backendQueue = make(chan *Backend, 10)
}

func getBackend() (*Backend, error) {
	select {
	case be := <-backendQueue:
		return be, nil
	case <-time.After(100 * time.Millisecond):
		be, err := net.Dial("tcp", "127.0.0.1:8081")
		if err != nil {
			return nil, err
		}

		return &Backend{
			Conn:   be,
			Reader: bufio.NewReader(be),
			Writer: bufio.NewWriter(be),
		}, nil
	}
}

func queueBackend(be *Backend) {
	select {
	case backendQueue <- be:
		// Backend re-enqueued safely, move on.
	case <-time.After(1 * time.Second):
		be.Close()
	}
}

func updateStats(req *http.Request, resp *http.Response) int64 {
	requestLock.Lock()
	defer requestLock.Unlock()

	bytes := requestBytes[req.URL.Path] + resp.ContentLength
	requestBytes[req.URL.Path] = bytes

	return bytes
}

func main() {
	// 1. Listen for connections forever.
	ln, err := net.Listen("tcp", ":8080")
	// add error checking
	if err != nil {
		log.Fatalf("Failed to listen: %s", err)
		return
	}
	for {

		//2. Accept connections.
		if conn, err := ln.Accept(); err == nil {

			go handleConnection(conn)

		}
	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		req, err := http.ReadRequest(reader)
		if err != nil {
			if err != io.EOF {
				log.Printf("Failed to read request: %s", err)
			}
			return
		}

		//		hostname := strings.Join(req.URL.Query()["hostname"], "")

		log.Printf("req.URL.Host: %s", req.Host)
		//		log.Printf("req: %s", req)
		//		log.Printf(req.Host)

		hostPortURL := req.URL
		var address string

		if hostPortURL.Opaque == "443" { //https访问
			address = hostPortURL.Scheme + ":443"
		} else { //http访问
			if strings.Index(hostPortURL.Host, ":") == -1 { //host不带端口， 默认80
				address = hostPortURL.Host + ":80"
			} else {
				address = hostPortURL.Host
			}
		}

		//		if len(hostname) == 0 {
		//			hostname = "baidu.com:80"
		//		} else {
		//			hostname += ":80"
		//		}

		be, err := net.Dial("tcp", address)
		if err != nil {
			log.Printf("Failed to connect to the web site: %s, err: %s", address, err)
			return
		}
		be_reader := bufio.NewReader(be)

		if err := req.Write(be); err == nil {
			//6. Read the response from the backend
			if resp, err := http.ReadResponse(be_reader, req); err == nil {
				bytes := updateStats(req, resp)
				resp.Header.Set("X-Bytes", strconv.FormatInt(bytes, 10))
				//				log.Printf("req: %s", resp)

				if err := resp.Write(conn); err == nil {
					//					log.Printf("req: %s", resp)
					log.Printf("%s: %d", req.URL.Path, resp.StatusCode)
				}

			}
		}

	}
}
