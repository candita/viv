// relay problem for Vivint 11/2017
// Relay needs to:
// - listen for tcp connections for a connected port
// - alert application when incoming connection is made
// - forward all traffic in both directions between registered app and requester
package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"strings"
)

const (
	RELAY_REQUEST = "relay:"
)

type connInfo struct {
	IpAddress string
	Port      string
}

// Handle the request to the relay server.  It is either a connection request or a relay setup request
func relay(appName string, conn net.Conn) {
	fmt.Printf("For %s, local: %s, remote: %s\n", appName, conn.LocalAddr(), conn.RemoteAddr())
	for {
		var bytes = make([]byte, 2048)
		numBytes, err := conn.Read(bytes)
		if err != nil {
			if err == io.EOF {
				continue
			}
			fmt.Printf("Error reading connection: %s\n", err.Error())
		} else {
			content := string(bytes[:numBytes])
			//fmt.Printf("INPUT: %s\n", content)

			// If it is a relay setup request call askRelay
			if strings.Contains(content, RELAY_REQUEST) {
				// Request should be formatted as relay:appname
				parts := strings.Split(content, ":")
				if len(parts) < 2 {
					conn.Write([]byte("Error - no application name specified\n"))
				} else {
					port := askRelay(parts[1])
					// Return a newline terminated message with the port
					conn.Write([]byte(":" + port + "\n"))
				}
			} else {
				// Otherwise it is a connection request, call askConnection
				askConnection(content, conn)
			}
		}
	}
}

func listen(appName, addr, port string, portChan chan string) {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Printf("Error on listen: %s\n", err.Error())
		return
	}
	defer ln.Close()
	fmt.Printf("Listening on %s...\n", ln.Addr())
	parts := strings.Split(ln.Addr().String(), ":")
	if len(parts) < 2 {
		fmt.Printf("Error in reading port from address: %s", ln.Addr())
		return
	}
	// Notify which port was obtained
	portChan <- parts[len(parts)-1]
	for {
		conn, err := ln.Accept()
		defer conn.Close()
		if err != nil {
			fmt.Println("Error on connection accept: %s", err.Error())
			return
		}
		go relay(appName, conn)
	}
}

// For an app asking for a relay, listen and return listening port
func askRelay(appName string) string {
	portChan := make(chan string)
	go listen(appName, "", "0", portChan)
	port := <-portChan
	return port
}

// Simulate an app asking for a connection
func askConnection(data string, conn net.Conn) {
	// Setup the tunnel between remote and the port for this app
	deliverTraffic(data, conn)
}

// Deliver traffic between the two endpoints
func deliverTraffic(data string, conn net.Conn) {
	//fmt.Printf("deliver -- local: %s, remote: %s\n", conn.LocalAddr(), conn.RemoteAddr())
	// Just write to endpoint
	conn.Write([]byte(data))
}

func main() {
	var port = flag.String("port", "8080", "Relay server listen port")
	flag.Parse()
	if *port == "" {
		*port = "8080"
	}
	ch := make(chan string)
	go listen("relay", "", *port, ch)
	_ = <-ch // empty it
	select {}
}
