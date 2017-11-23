***Build***
```
go build -o relay

cd echoserver

go build 
```
***Deploy***
```
../relay <port>
```
The port is optional but set to 8080 if omitted.

***Usage***

To configure your program to use the relay, follow this algorithm.  See echoserver/echoserver.go for an example of a program that follows it.
1. Make a tcp connection to the relay server using the port you deployed in the Deploy step above
2. Send a message containing "deadbeaffade" to the tcp connection, and read the response.  The response will contain a public port to which you can connect a second program.  Publish this port for your clients.
3. Send a message containing "badefeedafed" to the tcp connection, and read the response.  The response will contain a private port which you need to open and use as a tcp listener.
4. Accept connections from the tcp listener, and read/write to them.  Remove colons from any new content that you write, but do not remove the port number prefix and colon from the content.

Now when a client accesses the public port on your program, it is connected to the relay server, and the relay server sends/receives messages to/from your program.

To use the relay, note the relay's host address and deployed port.   
Use the relay's host and port as parameters when you start your program, for example:
./echoserver [host] [port]

To configure a client to use the relay, use the public port described in step 2 above, for example:
telnet localhost [publicPort]

You can now send traffic back and forth from your client program to your program via the relay.
