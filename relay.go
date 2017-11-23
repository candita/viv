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
	LISTEN_PORT   = "badefeedafed"
)

type route struct {
	Connection net.Conn // Configured connection
}

var (
	relayPort string
	routes    map[string]route  // Index is remote port
	relays    map[string]string // Index is remote port
	listeners map[string]string // Index is listener port
	returns   map[string]string // Index is local port
)

// Handle the request to the relay server.  It is either a connection request or a relay setup request
func relay(conn net.Conn) {
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
			//fmt.Printf("\nLocal %s, remote: %s\n", conn.LocalAddr(), conn.RemoteAddr())
			//fmt.Printf("INPUT: %s\n", content)

			// Store this route to write on later
			p, _ := getPort(conn.RemoteAddr())
			routes[p] = route{conn}

			// If it is a relay setup request call askRelay
			if strings.Contains(content, RELAY_REQUEST) {
				port := askRelay()
				if port == "none" {
					conn.Write([]byte("Error - no free relay ports"))
					return
				} else {
					// Remember where this relay request came from
					p, _ = getPort(conn.RemoteAddr())
					relays[port] = p
					// Send a newline terminated message with the :port
					conn.Write([]byte(":" + port + "\n"))
				}
				continue
				// If it is a listening port request, get a new port and write it
			} else if strings.Contains(content, LISTEN_PORT) {
				newPort := askListen()
				if newPort == "none" {
					conn.Write([]byte("Error - no free ports"))
					return
				}
				// Remember where this new port request goes to
				p, _ := getPort(conn.RemoteAddr())
				listeners[p] = newPort
				conn.Write([]byte("Listen:" + newPort + "\n"))
				continue
			} else {
				// Otherwise it is a client request or response, find saved connection or create new
				var newConn net.Conn
				// Check for a listener or control conn  made already for the client destination program

				p, _ := getPort(conn.LocalAddr())
				r, _ := getPort(conn.RemoteAddr())
				relayPort, okr := relays[p]
				if okr {
					listenerPort, ok := listeners[relayPort]
					if ok {
						//Save return addr for this conn
						returns[listenerPort] = r

						// dial it up and write to it
						newConn, err = net.Dial("tcp", ":"+listenerPort)
						if err != nil {
							fmt.Printf("Error creating conn for write to %s via %s: %s\n", p, listenerPort, err.Error())
						} else if newConn != nil {
							// Preface it with the return port
							newConn.Write([]byte(r + ":" + content))
							defer newConn.Close()
						} else {
							fmt.Printf("Error - connection was terminated, could not write %s\n", content)
						}
						continue
					} else {
						fmt.Printf("Error - no listener for relay %s\n", p)
						return
					}

				} else {
					// This is a response, not a request
					// Try to see if there is a routable port on content prefix
					parts := strings.Split(content, ":")
					if len(parts) > 1 {
						port := parts[0]
						content = strings.Replace(content, port+":", "", -1)
						// Find the conn to write to
						_, ok := routes[port]
						if ok {
							savedConn := routes[port].Connection
							savedConn.Write([]byte(content))
							continue
						} else {
							fmt.Printf("Error - connection missing for port %s\n", port)
							return
						}
					}
				}
			}
			// If all else fails
			conn.Write([]byte("Unroutable:" + content))
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

// For an app asking for a listen port for itself, find a free port
// TBD - make sure it doesn't get used before the app made it
func askListen() string {
	ln, err := net.Listen("tcp", ":0")
	if ln != nil {
		defer ln.Close() // Always close it, we won't need to listen in this server
	}
	if err != nil {
		fmt.Printf("Error on listen: %s\n", err.Error())
		return "none"
	}
	assignedPort, err := getPort(ln.Addr())
	if err != nil {
		fmt.Println(err.Error())
		return "none"
	}
	return assignedPort
}

// ./relay port
func main() {
	if len(os.Args) < 2 {
		relayPort = "8080"
	} else {
		relayPort = os.Args[1]
	}
	routes = make(map[string]route)
	relays = make(map[string]string)
	listeners = make(map[string]string)
	returns = make(map[string]string)
	ch := make(chan string)
	defer close(ch)
	go listen(relayPort, ch)
	_ = <-ch // empty it
	select {}
}
