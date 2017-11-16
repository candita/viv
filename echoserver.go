package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

const (
        RELAY_REQUEST = "relay"
)

// Try to listen on the assigned ip and port
func echo(ipport string){
	ln, err := net.Listen("tcp", ipport)
	if err != nil {
		fmt.Printf("Error on listen: %s\n", err.Error())
	//	os.Exit(1)
	}
	defer ln.Close()
	fmt.Printf("Listening on %s...\n", ln.Addr())
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error on connection accept: %s", err.Error())
			os.Exit(1)
		}
		defer conn.Close()
		fmt.Println("echoing...")
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
	if  err != nil {
		fmt.Println("Error dialing relay server: %v", err.Error())
		os.Exit(1)
	}
	// Send a relay request, specifying the app name
	fmt.Fprintf(conn, RELAY_REQUEST+":echoserver")
	ipport, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading relay request results: %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Printf("established relay address: %s\n",ipport)

	go echo(ipport)

}

