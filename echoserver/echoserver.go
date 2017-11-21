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

func copy(listenPort, relayPort string) {
	fmt.Printf("Got port: %s\n", listenPort)
	ln, err := net.Listen("tcp", ":"+listenPort)
	if err != nil {
		fmt.Printf("Error on listen: %s\n", err.Error())
		if ln != nil {
			ln.Close()
		}
		return
	}
	fmt.Printf("Listening on: %s\n", listenPort)
	defer ln.Close()
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
		newConn, err := net.Dial("tcp", ":"+relayPort)
		if err != nil {
			fmt.Printf("Error dialing relay server: %s\n", err.Error())
			if newConn != nil {
				newConn.Close()
			}
			return
		}
		defer newConn.Close()
		fmt.Println("Writing " + content)
		newConn.Write([]byte(content))
	}
}

func main() {
	var host, relayPort string
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
	ports := strings.Split(contents, ":")
	if len(ports) < 2 {
		fmt.Println("Error establishing connection, ports unassigned")
		os.Exit(1)
	}
	fmt.Printf("established relay address: %s\n", ports[1])

	for {
		copy(ports[0], relayPort)
	}
	select {}
}
