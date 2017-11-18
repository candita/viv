*Build*

go build -o relay
cd echoserver
go build 

*Deploy*

../relay <port>

*Usage*

To configure your program to use the relay, send a message containing "deadbeaffade" to the relay's host 
and port, and read the response.  The response will contain a port to which you can connect a second program.
The relay server has set up a relay between your program, itself, and that port.  When you connect a second 
program to this port, you are talking to your program through the relay. 

Example code that you add to your program:   
```// Dial the relay host/port
conn, err := net.Dial("tcp", host+":"+port)
// Send the message
fmt.Fprintf(conn, "deadbeaffade")
// Read the relay port
port, err := bufio.NewReader(conn).ReadString('\n')
// Print the port and get on with the rest of your program
```

To use the relay, note the relay's host address and configured port.   
Use the relay's host and port as parameters when you start your program, for example:
./echoserver <host> <port>

Note the port that was printed and connect your second program to this port.  You can now send traffic
back and forth from your second program to your program via the relay.

