// relay problem for Vivint 11/2017
// Relay needs to:
// - listen for tcp connections for a connected port
// - alert application when incoming connection is made
// - forward all traffic in both directions between registered app and requester
package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

const (
	RELAY_REQUEST = "deadbeaffade"
)

type route struct {
	Connection net.Conn // Configured connection
}

var (
	relayPort string
	routes    map[string]route // Index is relay port
)

// Handle the request to the relay server.  It is either a connection request or a relay setup request
func relay(conn net.Conn) {
	//fmt.Printf("local: %s, remote: %s\n", conn.LocalAddr(), conn.RemoteAddr())
	for {
		var bytes = make([]byte, 2048)
		numBytes, err := conn.Read(bytes)
		if numBytes == 0 {
			// Connection was gracefully closed, exit
			return
		}
		if err != nil {
			if err == io.EOF {
				continue
			}
			fmt.Printf("Error reading connection: %s\n", err.Error())
			return
		} else {
			content := string(bytes[:numBytes])
			fmt.Printf("Local %s, remote: %s\n", conn.LocalAddr(), conn.RemoteAddr())
			fmt.Printf("INPUT: %s\n", content)

			// If it is a relay setup request call askRelay
			if strings.Contains(content, RELAY_REQUEST) {
				port := askRelay()
				if port == "none" {
					conn.Write([]byte("Error - no free ports"))
					return
				} else {
					// Send a newline terminated message with the :port
					conn.Write([]byte(":" + port + "\n"))
				}
			} else {
				// Otherwise it is a connection request, find saved connection or create new
				savedConn, ok := routes[conn.RemoteAddr().String()]
				if !ok {
					// If this is the first contact from this client, set a new connection
					newPort := askRelay()
					// Write the new port to the old connection
					conn.Write([]byte("Listen:" + newPort))
					// May want to sleep here

				} else {
					conn = savedConn.Connection // write to the connection stored for this
				}
				conn.Write([]byte(content))
				fmt.Printf("Wrote %s to %s\n", content, conn.RemoteAddr().String())
			}
		}
	}
}

// Get the port from the Addr
func getPort(addr net.Addr) (string, error) {
	parts := strings.Split(addr.String(), ":")
	if len(parts) < 2 {
		return "", fmt.Errorf("Error in reading port from address: " + addr.String())
	}
	return parts[len(parts)-1], nil
}

func listen(port string, portChan chan string) {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Printf("Error on listen: %s\n", err.Error())
		if ln != nil {
			ln.Close()
		}
		return
	}
	defer ln.Close()
	fmt.Printf("Listening on %s...\n", ln.Addr())
	assignedPort, err := getPort(ln.Addr())
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// Notify which port was obtained (in case it was passed as zero)
	portChan <- assignedPort
	for {
		conn, err := ln.Accept()
		defer conn.Close()
		if err != nil {
			fmt.Println("Error on connection accept: %s", err.Error())
			if conn != nil {
				conn.Close()
			}
			return
		}
		// Save this route for lookup later
		routes[conn.RemoteAddr().String()] = route{conn}
		go relay(conn)
	}
}

// For an app asking for a relay, listen and return listening port
func askRelay() string {
	portChan := make(chan string)
	defer close(portChan)
	go listen("0", portChan)
	// In case there is no port available, timeout
	select {
	case port := <-portChan:
		return port
	case <-time.After(time.Second * 1):
		return "none"
	}
}

// ./relay port
func main() {
	if len(os.Args) < 2 {
		relayPort = "8080"
	} else {
		relayPort = os.Args[1]
	}
	routes = make(map[string]route)
	ch := make(chan string)
	defer close(ch)
	go listen(relayPort, ch)
	_ = <-ch // empty it
	select {}
}
