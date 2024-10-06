package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
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

type CustomResponse struct {
	Headers        map[string]string
	HttpVersion    string
	HttpStatusName string
	Body           []byte
	HttpStatus     int16
}

var (
	_ = net.Listen
	_ = os.Exit
)

func main() {
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

// Bit ugly but...
func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	request := parseRequest(reader)
	paths := strings.Split(request.Path, "/")
	var response CustomResponse
	response.HttpVersion = request.HttpVersion
	response.Headers = make(map[string]string)
	response.Headers["Content-Type"] = "text/plain"

	var responseString string

	switch {
	case paths[1] == "user-agent":
		userAgent := request.Headers["User-Agent"]
		// oldresponse := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)
		response.HttpStatus = 200
		response.HttpStatusName = "OK"
		response.Headers["Content-Length"] = strconv.Itoa(len(userAgent))
		response.Body = []byte(userAgent)
	case paths[1] == "echo":
		response.HttpStatus = 200
		response.HttpStatusName = "OK"
		if strings.Contains(request.Headers["Accept-Encoding"], "gzip") {
			var b bytes.Buffer
			enc := gzip.NewWriter(&b)
			enc.Write([]byte(paths[2]))
			enc.Close()
			response.Headers["Content-Encoding"] = "gzip"
			response.Headers["Content-Length"] = strconv.Itoa(len(b.String()))
			response.Body = b.Bytes()
		} else {
			response.Headers["Content-Length"] = strconv.Itoa(len(paths[2]))
			response.Body = []byte(paths[2])
		}
	case paths[1] == "files":
		handleFiles(request, &response)
	case request.Path == "/":
		response.HttpStatus = 200
		response.HttpStatusName = "OK"
	default:
		response.HttpStatus = 404
		response.HttpStatusName = "Not Found"
	}
	responseString = parseResponse(&response)
	conn.Write([]byte(responseString))
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

func handleFiles(request *CustomRequest, response *CustomResponse) {
	switch request.Method {
	case "GET":
		dir := os.Args[2]
		fileName := strings.TrimPrefix(request.Path, "/files/")
		fileString := fmt.Sprintf("%s%s", dir, fileName)
		file, err := os.ReadFile(fileString)
		if err != nil {
			response.HttpStatus = 404
			response.HttpStatusName = "Not Found"
		} else {
			response.HttpStatus = 200
			response.HttpStatusName = "OK"
			response.Headers["Content-Length"] = strconv.Itoa(len(file))
			response.Headers["Content-Type"] = "application/octet-stream"
			response.Body = file
		}
	case "POST":
		dir := os.Args[2]
		fileName := strings.TrimPrefix(request.Path, "/files/")
		fileString := fmt.Sprintf("%s%s", dir, fileName)
		content := []byte(request.Body)
		err := os.WriteFile(fileString, content, 0644)
		if err != nil {
			fmt.Println(err)
			response.HttpStatus = 200
			response.HttpStatusName = "OK"
			break
		} else {
			response.HttpStatus = 201
			response.HttpStatusName = "Created"
		}
	default:
		response.HttpStatus = 405
		response.HttpStatusName = "Not Allowed"
	}
}

func parseResponse(response *CustomResponse) string {
	var httpStatus, headers string
	httpStatus = fmt.Sprintf("%s %d %s", response.HttpVersion, response.HttpStatus, response.HttpStatusName)

	for left, right := range response.Headers {
		headers = headers + left + ": " + right + "\r\n"
	}

	body := response.Body

	return fmt.Sprintf("%s\r\n%s\r\n%s", httpStatus, headers, body)
}
