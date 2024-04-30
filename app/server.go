package main

import (
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
	"path"
	"strings"
)

const (
	PORT             = "4221"
	VERSION          = "HTTP/1.1"
	CRLF             = "\r\n"
	MAX_REQUEST_SIZE = 1024
	OK               = "200 OK"
	NOT_FOUND        = "404 Not Found"
	CREATED          = "201 Created"
	NOT_ALLOW        = "405 Method Not Allowed"
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
	Body    []byte
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
	case 405:
		response.Status = NOT_ALLOW
	case 201:
		response.Status = CREATED
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
		fmt.Println(string(line))
		if i == 0 {
			fmt.Sscanf(string(line), "%s %s %s", &req.Method, &req.Path, &req.Version)
		} else if i == len(lines)-1 {
			req.Body = line
		} else {
			header_elem := bytes.Split(line, []byte(": "))
			if len(header_elem) == 2 {
				key := string(header_elem[0])
				value := string(header_elem[1])
				req.Headers[key] = value
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
	var directory string
	flag.StringVar(&directory, "directory", "", "")
	flag.Parse()
	fmt.Println("File Directory:", directory)
	if directory != "" {
		if _, err := os.Stat(directory); os.IsNotExist(err) {
			fmt.Println("Passed Directory does not exist", err)
			os.Exit(1)
		}
	}

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
		go HandleConnection(conn, directory)
	}
}

func HandleConnection(conn net.Conn, directory string) {
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

	} else if strings.HasPrefix(request.Path, "/files/") {

		fileName := strings.TrimPrefix(request.Path, "/files/")
		filePath := path.Join(directory, fileName)

		switch request.Method {

			case "GET":
				file, err := os.Open(filePath)
				if err != nil {
					conn.Write(NewHTTPResponse(404, headers, body).ToBytes())
				}
				defer file.Close()
				body, err = os.ReadFile(filePath)
				if err != nil {
					fmt.Println("Error reading file: ", err.Error())
					os.Exit(1)
				}
				headers["Content-Type"] = "application/octet-stream"
				conn.Write(NewHTTPResponse(200, headers, body).ToBytes())

			case "POST":
				fileName := strings.TrimPrefix(request.Path, "/files/")
				filePath := path.Join(directory, fileName)
				file, err := os.Create(filePath)
				if err != nil {
					fmt.Println("Error creating file: ", err.Error())
					os.Exit(1)
				}
				defer file.Close()
				_, err = file.Write(request.Body)
				if err != nil {
					fmt.Println("Error writing to file: ", err.Error())
					os.Exit(1)
				}
				headers["Content-Type"] = "application/octet-stream"
				conn.Write(NewHTTPResponse(201, headers, body).ToBytes())
				
			default:
				conn.Write(NewHTTPResponse(405, headers, body).ToBytes())
		}

	} else {
		conn.Write(NewHTTPResponse(404, headers, body).ToBytes())
	}

}
