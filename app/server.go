package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

type CustomRequest struct {
	Method      string
	Path        string
	Headers     map[string]string
	HttpVersion string
	Body        string
	UserAgent   string
}

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

	reader := bufio.NewReader(conn)

	if err != nil {
		fmt.Println("There was an error: ", err.Error())
	}

	request := parseRequest(reader)

	paths := strings.Split(request.Path, "/")

	if request.Path == "/user-agent" {
		userAgent := request.Headers["User-Agent"]
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)
		conn.Write([]byte(response))
	} else if paths[1] == "echo" {
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(paths[2]), paths[2])
		conn.Write([]byte(response))
	} else if request.Path == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}

func parseRequest(reader *bufio.Reader) *CustomRequest {
	// Parse request information i.e. request method, url, http version
	requestInfo, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("There was an error requesting info: ", err.Error())
	}

	urlParts := strings.Split(requestInfo, "\r\n")

	parsedRequest := CustomRequest{}
	headers := make(map[string]string)

	methodPathVersion := strings.Split(urlParts[0], " ")

	parsedRequest.Method = methodPathVersion[0]
	parsedRequest.Path = methodPathVersion[1]
	parsedRequest.HttpVersion = methodPathVersion[2]

	fmt.Println("Calculating headers")
	// Parse headers
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("There was an error requesting info: ", err.Error())
		}

		line = strings.Trim(line, "\n\r")

		// End of headers
		if line == "" {
			break
		}

		name, value, found := strings.Cut(line, ": ")
		if !found {
			fmt.Println("Wrong header format", err.Error())
		}
		headers[name] = value
	}

	parsedRequest.Headers = headers

	return &parsedRequest
}
