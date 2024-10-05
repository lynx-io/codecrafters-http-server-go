package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var (
	_ = net.Listen
	_ = os.Exit
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	port := 4221
	address := "0.0.0.0:" + strconv.Itoa(port)

	l, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("Failed to bind to port:", port)
		os.Exit(1)
	}
	fmt.Println("Listening on: ", address)

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	status, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("There was an error: ", err.Error())
	}

	request := strings.Fields(status)

	url := request[1]

	paths := strings.Split(url, "/")

	if paths[1] == "echo" {
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(paths[2]), paths[2])
		conn.Write([]byte(response))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}
