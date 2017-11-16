// relay problem for Vivint 11/2017
// Relay needs to:
// - listen for tcp connections for a registered application
// - alert application when incoming connection is made
// - forward all traffic in both directions between registered app and requester
package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

const (
	RELAY_REQUEST = "relay:"
)

type connInfo struct {
	IpAddress string
	Port      string
}

var (
	// For each app, hold the connInfo (ip address and port)
	appMap map[string]connInfo
	// The list of assignable ports (later, read from config file)
	availablePorts [4]string
	usedPorts      [4]string
)

func init() {

	// For now, just use a hard-coded list
	availablePorts = [...]string{"8081", "8082", "8083", "8084"}
	appMap = make(map[string]connInfo)
}

// Handle the request to the relay server.  It is either a connection request or a relay setup request
func relay(appName string, conn net.Conn) {
	//for {

		fmt.Printf("local: %s, remote: %s\n", conn.LocalAddr(), conn.RemoteAddr())
		var bytes = make([]byte, 2048)
		numBytes, err := conn.Read(bytes)
		if err != nil {
			if err == io.EOF {
				return //continue
			}
			fmt.Printf("Error reading connection: %s\n", err.Error())
		} else {
			content := string(bytes[:numBytes])
			fmt.Printf("INPUT: %s\n", content)

			// If it is a relay setup request call askRelay
			if  strings.Contains(content, RELAY_REQUEST){
				// Request should be formatted as relay:appname
				parts :=  strings.Split(content,":")
				if len(parts) < 2 {
					conn.Write([]byte("Error: no application name specified\n"))
				} else {
					port,err := askRelay(parts[1])
					if  err != nil {
						conn.Write([]byte(fmt.Sprintf("Error requesting relay %s\n",err.Error())))
					}
					// Return a newline terminated message with the port
					conn.Write([]byte(":" + port + "\n"))
				}
			} else {
				// Otherwise it is a connection request, call askConnection with the message content
				if err != nil {
					fmt.Printf("Error granting connection request: %s\n", err.Error())
				} else {
					askConnection(conn)
				}
			}
		}
	//}
}

func listen(appName, addr, port string) {
	ln, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Printf("Error on listen: %s\n", err.Error())
		return
	}
	defer ln.Close()
	ci, there := appMap[appName]
	if !there {
		appMap[appName] = connInfo{addr, port}
	} else {
		if ci.Port != port {
			fmt.Println("Already listening: " + appName)
			return
		}
	}
	fmt.Printf("Listening on %s...\n", ln.Addr())
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error on connection accept: %s", err.Error())
			return
		}
		go relay(appName, conn)
	}
}

// Return the next port to assign
func getPort(appName string) (string, error) {
	if appName != "" {
		ci, exists := appMap[appName]
		if exists {
			return ci.Port, nil
		}
	}
	for i, port := range usedPorts {
		if port == "" {
			usedPorts[i] = "used"
			return availablePorts[i], nil
		}
	}
	return "", fmt.Errorf("Fatal error - all ports in use")
}

// For an app asking for a relay, return the port
func askRelay(appName string) (string, error) {
	port, err := getPort(appName)
	if err != nil {
		fmt.Printf(err.Error())
		return "", err
	}
	go listen(appName, "", port)
	return port, nil
}

// Simulate an app asking for a connection
func askConnection(conn net.Conn) {
	// Setup the tunnel between remote and the port for this app
	go deliverTraffic(conn)
}

func read(remoteConn net.Conn, ch chan []byte){
		bytes := make([]byte, 2048)
		var content string
		remoteBytes, err := remoteConn.Read(bytes)
		if err != nil {
			fmt.Printf("Error reading connection: %s\n", err.Error())
			content = string(bytes[:remoteBytes])
		} else {
			fmt.Printf("Read %s\n", content)
		}
		ch <- bytes
}

func write(localConn net.Conn, ch chan []byte){
		bytes := <- ch
		localBytes, err := localConn.Write(bytes)
		if err != nil {
			fmt.Printf("Error writing connection: %s\n", err.Error())
		} else {
			fmt.Printf("Wrote %s, %d bytes\n",bytes[:localBytes], localBytes)
		}
}

// Deliver traffic between the two endpoints
func deliverTraffic(conn net.Conn){
	remoteConn, err := net.Dial("tcp", conn.RemoteAddr().String())
	if err != nil {
		fmt.Printf("Error connecting to remote %v: %s\n", conn.RemoteAddr(), err.Error())
		return
	}
	localConn, err := net.Dial("tcp", conn.RemoteAddr().String())
	if err != nil {
		fmt.Printf("Error connecting to local %v: %s\n", conn.RemoteAddr(), err.Error())
		return
	}
	// Now read and write between the two endpoints
	for {
		var ch = make(chan []byte)
		defer close(ch)
		go read(remoteConn, ch)
		go write(localConn, ch)
	}
}

func main() {
	var port = flag.String("port", "8080", "Relay server listen port")
	flag.Parse()
	if *port == "" {
		*port = "8080"
	}
	go listen("relay", "", *port)
	time.Sleep(2)
	select{}
}

