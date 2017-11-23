package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

const (
	MY_RELAY_REQUEST = "deadbeaffade"
	MY_LISTEN_PORT   = "badefeedafed"
)

func copy(oldConn net.Conn) {
	// Listen for a message about a new connection
	l, err := getConnection(oldConn)
	if err != nil {
		if l != nil {
			l.Close()
		}
		fmt.Printf("Error getting a new connection: %s\n", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if conn != nil {
			defer conn.Close()
		}
		if err != nil {
			fmt.Println("Error on connection accept: %s", err.Error())
			return
		}
		// Read
		var bytes = make([]byte, 2048)
		numBytes, err := conn.Read(bytes)
		if err != nil {
			if err == io.EOF {
				continue
			}
			fmt.Printf("Error reading connection: %s\n", err.Error())
			return
		}
		if numBytes == 0 {
			// Connection was gracefully closed, exit
			return
		}

		// Write it back to the relayserver
		content := string(bytes[:numBytes])
		oldConn.Write([]byte(content)) // only this can be seen on the relayserver
	}
}

// Return a new relay port
func relayRequest(conn net.Conn) (string, error) {
	// Send a relay request to the given conn
	fmt.Fprintf(conn, MY_RELAY_REQUEST)
	contents, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return "", err
	}
	// Should receive back the relay address but could receive an error
	if strings.Contains(contents, "Error") {
		fmt.Println(contents)
		return "", fmt.Errorf(contents)
	}
	return contents, nil

}

// Return a new connection
func getConnection(conn net.Conn) (newConn net.Listener, err error) {
	var port, contents string
	// Explicitly ask for the connection port
	fmt.Fprintf(conn, MY_LISTEN_PORT)
	contents, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return nil, err
	}
	// Should receive back the Listen:port message
	if strings.Contains(contents, "Listen:") {
		port = strings.Replace(contents, "Listen:", "", -1)
		port = strings.TrimRight(port, "\n")
		newConn, err = net.Listen("tcp", ":"+port)
		if err != nil {
			return nil, err
		}
		return newConn, nil
	}
	return nil, fmt.Errorf("No Listen port provided")
}

var host, relayPort, publicPort string

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: echoserver <hostname> <portnum>")
		os.Exit(1)
	} else {
		host = os.Args[1]
		relayPort = os.Args[2]
	}

	// Send a message to the host:port asking for a relay host:port
	conn, err := net.Dial("tcp", host+":"+relayPort)
	if err != nil {
		fmt.Printf("Error dialing relay server: %s\n", err.Error())
		if conn != nil {
			conn.Close()
		}
		os.Exit(1)
	}
	defer conn.Close()

	// Send a relay request to get a public port
	publicPort, err = relayRequest(conn)
	if err != nil {
		fmt.Printf("Error reading relay request results for public port: %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Printf("established relay address: %s\n", publicPort)

	go copy(conn)
	select {} // wait here
}
