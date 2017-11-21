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
	FromPort  string // The client that tries to access program
	RelayPort string // The assigned relay port
	ToPort    string // The program
}

var (
	relayPort     string
	assignedPorts map[string]string
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
				} else {
					// Record the assigned port as a connection to conn.RemoteAddr
					assignedPorts[port], _ = getPort(conn.RemoteAddr())
					fmt.Printf("Assigned port: %s\n", assignedPorts[port])
					// Return a newline terminated message with the remotePort:port
					conn.Write([]byte(assignedPorts[port] + ":" + port + "\n"))
				}
			} else {
				// Otherwise it is a connection request, deliver traffic

				// If available, get the return port
				returnPort, _ := getPort(conn.LocalAddr())
				port, ok := assignedPorts[returnPort]
				if !ok {
					fmt.Println("Error - could not find a valid destination")
					return
				}
				fmt.Printf("Return port: %s\n", port)
				// Dial return port
				newConn, err := net.Dial("tcp", ":"+port)
				if err != nil {
					fmt.Printf("Error dialing relayed program: %s\n", err.Error())
					if newConn != nil {
						newConn.Close()
					}
					return
				}
				defer newConn.Close()

				// Write to the connection
				newConn.Write([]byte(content))
				fmt.Printf("Wrote %s to %s\n", content, port)
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
	assignedPorts = make(map[string]string)
	ch := make(chan string)
	defer close(ch)
	go listen(relayPort, ch)
	_ = <-ch // empty it
	select {}
}
