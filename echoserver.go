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
		if strings.Contains(contents, "Error") {
			fmt.Println(contents)
		} else {
			fmt.Printf("Error reading relay request port assignments\n")
		}
		os.Exit(1)
	}
	fmt.Printf("established relay address: %s\n", ports[1])

	for {
		io.Copy(conn, conn)
	}
}
