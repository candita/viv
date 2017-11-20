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
)

func copy(conn net.Conn) {
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
		// Write
		content := string(bytes[:numBytes])
		conn.Write([]byte(content))
	}
}

func main() {
	var host, port string
	if len(os.Args) < 3 {
		fmt.Println("Usage: echoserver <hostname> <portnum>")
		os.Exit(1)
	} else {
		host = os.Args[1]
		port = os.Args[2]
	}

	// Send a message to the host:port asking for a relay host:port
	conn, err := net.Dial("tcp", host+":"+port)
	if err != nil {
		fmt.Printf("Error dialing relay server: %s\n", err.Error())
		if conn != nil {
			conn.Close()
		}
		os.Exit(1)
	}
	defer conn.Close()
	// Send a relay request
	fmt.Fprintf(conn, MY_RELAY_REQUEST)
	contents, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading relay request results: %s\n", err.Error())
		os.Exit(1)
	}
	// Should receive back the relay address but could receive an error
	if strings.Contains(contents, "Error") {
		fmt.Println(contents)
		os.Exit(1)
	}
	fmt.Printf("established relay address: %s\n", contents)

	go copy(conn)
}
