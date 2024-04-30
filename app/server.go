package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
)

const (
	PORT = "4221"
	CRLF = "\r\n"
	MAX_REQUEST_SIZE = 1024
)

type HTTPResponse struct {
	Version string
	Status  string
}

type HTTPRequest struct {
	Method  string
	Path    string
	Version string
}

func NewHTTPResponse(code int) *HTTPResponse {
	switch code {
	case 200:
		return &HTTPResponse{
			Version: "HTTP/1.1",
			Status:  "200 OK",
		}
	case 404:
		return &HTTPResponse{
			Version: "HTTP/1.1",
			Status:  "404 Not Found",
		}
	}
	return &HTTPResponse{
		Version: "HTTP/1.1",
		Status:  "500 Internal Server Error",
	}
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
		// Headers: make(map[string]string),
	}
	lines := bytes.Split(buf, []byte(CRLF))
	fmt.Sscanf(string(lines[0]), "%s %s %s", &req.Method, &req.Path, &req.Version)

	return req, err
}

func (r *HTTPResponse) ToBytes() (response []byte) {
	response = []byte{}
	response = append(response, fmt.Sprintf("%s %s%s", r.Version, r.Status, CRLF)...)
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
		request, err := NewHTTPRequest(conn)
		if err != nil {
			fmt.Println("Error parsing request: ", err.Error())
			os.Exit(1)
		}
		switch request.Path {
		case "/":
			conn.Write(NewHTTPResponse(200).ToBytes())
		default:
			conn.Write(NewHTTPResponse(404).ToBytes())
		}
		conn.Close()
	}
}
