package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

const (
	RELAY_REQUEST = "relay"
)

// Echo
func echo(conn net.Conn) {
	fmt.Printf("echoing between... %v and %v\n", conn.LocalAddr(), conn.RemoteAddr())
	for {
		// Echo all incoming data
		io.Copy(conn, conn)
	}
}

func main() {
	var port = flag.String("port", "", "port number")
	var host = flag.String("host", "127.0.0.1", "host name")
	flag.Parse()
	if *port == "" {
		fmt.Println("Usage: echoserver -host [hostname] -port [portnum]")
		os.Exit(1)
	}
	if *host == "" {
		*host = "127.0.0.1"
	}

	// Send a message to the host:port asking for a relay host:port
	conn, err := net.Dial("tcp", *host+":"+*port)
	if err != nil {
		fmt.Println("Error dialing relay server: %v", err.Error())
		os.Exit(1)
	}
	defer conn.Close()
	// Send a relay request, specifying the app name
	fmt.Fprintf(conn, RELAY_REQUEST+":echoserver")
	contents, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading relay request results: %s\n", err.Error())
		os.Exit(1)
	}
	// Should receive back the relayed-port:echoserver-port
	ports := strings.Split(contents, ":")
	if len(ports) < 2 {
		fmt.Printf("Error reading relay request port assignments\n")
		os.Exit(1)
	}
	fmt.Printf("established relay address: %s\n", ports[1])

	go echo(conn)
}
