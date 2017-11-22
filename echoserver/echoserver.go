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
	conn, err := getConnection(oldConn)
	if err != nil {
		if conn != nil {
			conn.Close()
		}
		fmt.Printf("Error getting a new connection: %s\n", err.Error())
		os.Exit(1)
	}
	defer conn.Close()
	for {
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
		fmt.Printf("Local %s, remote: %s\n", conn.LocalAddr(), conn.RemoteAddr())
		fmt.Println("Reading " + string(bytes[:numBytes]))

		// Write
		content := string(bytes[:numBytes])
		// See if there is a return destination "xxx:content"
		//parts := strings.Split(content, ":")
		//if len(parts) > 1 {
		//relayPort = parts[0]
		//content = parts[1]
		//}

		fmt.Println("Writing " + content)
		conn.Write([]byte("echoed " + content))
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
func getConnection(conn net.Conn) (newConn net.Conn, err error) {
	var port, contents string
	// Explicitly ask for the connection port
	fmt.Fprintf(conn, MY_LISTEN_PORT)
	contents, err = bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return nil, err
	}
	// Should receive back the Listen:port message
	if strings.Contains(contents, "Listen:") {
		port = strings.Replace(contents, "Listen", "", -1)
		fmt.Printf("Received port: %s\n", port)
		newConn, err = net.Dial("tcp", host+port)
		if err != nil {
			return nil, err
		}
		return newConn, nil
	}
	return nil, fmt.Errorf("No Listen port provided")
}

var host, relayPort string

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
	publicPort, err := relayRequest(conn)
	if err != nil {
		fmt.Printf("Error reading relay request results for public port: %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Printf("established relay address: %s\n", publicPort)

	//for {
	go copy(conn)
	//}
	select {}
}
