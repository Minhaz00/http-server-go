package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	PORT             = "4221"
	VERSION          = "HTTP/1.1"
	CRLF             = "\r\n"
	MAX_REQUEST_SIZE = 1024
	OK               = "200 OK"
	NOT_FOUND        = "404 Not Found"
)

type HTTPResponse struct {
	Version string
	Status  string
	Headers map[string]string
	Body    []byte
}

type HTTPRequest struct {
	Method  string
	Path    string
	Version string
	Headers map[string]string
}

func NewHTTPResponse(code int, headers map[string]string, body []byte) (response *HTTPResponse) {
	response = &HTTPResponse{
		Version: VERSION,
		Headers: headers,
		Body:    body,
	}
	response.Headers["Content-Length"] = fmt.Sprintf("%d", len(body))

	switch code {
	case 200:
		response.Status = OK
	case 404:
		response.Status = NOT_FOUND
	}
	return response
}

func NewHTTPRequest(conn net.Conn) (req *HTTPRequest, err error) {
	buf := make([]byte, MAX_REQUEST_SIZE)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading request: ", err.Error())
		return req, err
	}

	buf = buf[:n]
	req = &HTTPRequest{
		Headers: make(map[string]string),
	}
	lines := bytes.Split(buf, []byte(CRLF))

	for i, line := range lines {
		if i == 0 {
			fmt.Sscanf(string(line), "%s %s %s", &req.Method, &req.Path, &req.Version)
		} else {
			header_elem := bytes.Split(line, []byte(": "))
			if len(header_elem) == 2 {
				key := string(header_elem[0])
				value := string(header_elem[1])
				req.Headers[key] = value
			} else {
				fmt.Println("Error parsing header: ", string(line))
			}
		}
	}

	fmt.Printf("Received request: %+v\n", req)

	return req, err
}

func (r *HTTPResponse) ToBytes() (response []byte) {
	response = []byte{}
	response = append(response, fmt.Sprintf("%s %s%s", r.Version, r.Status, CRLF)...)

	for k, v := range r.Headers {
		response = append(response, fmt.Sprintf("%s: %s%s", k, v, CRLF)...)
	}
	response = append(response, CRLF...)
	response = append(response, fmt.Sprintf("%s%s", r.Body, CRLF)...)

	response = append(response, CRLF...)
	return response
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", fmt.Sprintf(":%s", PORT))
	if err != nil {
		fmt.Printf("Error listening on port %s: %s\n", PORT, err.Error())
		os.Exit(1)
	}

	fmt.Printf("Listening on port %s\n", PORT)
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go HandleConnection(conn)
	}
}


func HandleConnection(conn net.Conn) {
	defer conn.Close()
	request, err := NewHTTPRequest(conn)
	if err != nil {
        fmt.Println("Error parsing request: ", err.Error())
		os.Exit(1)
    }

	headers := make(map[string]string)
	body := []byte{}

	if request.Path == "/" {
		conn.Write(NewHTTPResponse(200, headers, body).ToBytes())
	} else if strings.HasPrefix(request.Path, "/echo/") {
		body = []byte(strings.TrimPrefix(request.Path, "/echo/"))
		headers["Content-Type"] = "text/plain"
		conn.Write(NewHTTPResponse(200, headers, body).ToBytes())
	} else if request.Path == "/user-agent" {
		body = []byte(request.Headers["User-Agent"])
		headers["Content-Type"] = "text/plain"
		conn.Write(NewHTTPResponse(200, headers, body).ToBytes())
	} else {
		conn.Write(NewHTTPResponse(404, headers, body).ToBytes())
	}

}
