package main

import (
	"bufio"
	"fmt"
	"io"
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

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	request := parseRequest(reader)
	paths := strings.Split(request.Path, "/")

	if request.Path == "/user-agent" {
		userAgent := request.Headers["User-Agent"]
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)
		conn.Write([]byte(response))
		return
	} else if paths[1] == "echo" {
		encoding := ""
		fmt.Println(request.Headers)
		if strings.Contains(request.Headers["Accept-Encoding"], "gzip") {
			encoding = "Content-Encoding: gzip\r\n"
		}
		response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n%s\r\n%s", len(paths[2]), encoding, paths[2])
		conn.Write([]byte(response))
		return
	} else if paths[1] == "files" {
		switch request.Method {
		case "GET":
			dir := os.Args[2]
			fileName := strings.TrimPrefix(request.Path, "/files/")
			fileString := fmt.Sprintf("%s%s", dir, fileName)
			file, err := os.ReadFile(fileString)
			if err != nil {
				fmt.Println(err)
				conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
				return
			}
			response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(file), file)
			conn.Write([]byte(response))
			return
		case "POST":
			dir := os.Args[2]
			fileName := strings.TrimPrefix(request.Path, "/files/")
			fileString := fmt.Sprintf("%s%s", dir, fileName)
			content := []byte(request.Body)
			fmt.Println("Writing file in: ", fileString)
			fmt.Println("Content: ", request.Body)
			err := os.WriteFile(fileString, content, 0644)
			if err != nil {
				fmt.Println(err)
				conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
				return
			}
			response := "HTTP/1.1 201 Created\r\n\r\n"
			conn.Write([]byte(response))
			return
		}
		conn.Write([]byte("HTTP/1.1 405 Not Allowed\r\n\r\n"))
		return
	} else if request.Path == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		return
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		return
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

	// Read the body if there's a Content-Length header
	if contentLength, ok := parsedRequest.Headers["Content-Length"]; ok {
		length, err := strconv.Atoi(contentLength)
		if err == nil {
			body := make([]byte, length)
			_, err := io.ReadFull(reader, body)
			if err != nil {
				fmt.Println("Error reading body:", err)
			}
			parsedRequest.Body = string(body)
		}
	}

	if err != nil {
		fmt.Println("There was an error parsing body ", err.Error())
	}

	return &parsedRequest
}
