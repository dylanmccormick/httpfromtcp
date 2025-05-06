package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
	"strconv"

	"github.com/dylanmccormick/httpfromtcp/internal/headers"
)

type requestState int 

const( 
	requestStateParsingRequest requestState = iota
	requestStateParsingHeader
	requestStateParsingBody
	requestStateError
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	RawRequest string
	RequestState requestState
	Headers headers.Headers
	Body []byte
}

type RequestLine struct {
	HttpVersion string
	RequestTarget string
	Method string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, 8)
	r := &Request{
		RequestState:requestStateParsingRequest,
		Headers: make(headers.Headers),
		Body: []byte(""),
	}
	for r.RequestState != requestStateDone {
		n,err := reader.Read(buf)
		if n == 0 {
			r.RequestState = requestStateDone
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		_, err = r.parse(buf[:n])

		if err != nil {
			return nil, err
		}
	}
	contentLength, err := r.Headers.Get("content-length")
	if err == nil {
		if contentLength == "" {
			return r, nil 
		}
		length, err := strconv.Atoi(contentLength)
		if err != nil {
			return r, err
		}
		if len(r.Body) < length{
			return r, fmt.Errorf("Body provided (len: %d) shorter than content-length (len: %d).\n", len(r.Body), length)

		}
	}

	return r, nil
}

func (r *Request) parse(data []byte) (int, error){
	bytesParsed := 0
	count := 0 
	for {
		bytes := 0
		var err error
		switch r.RequestState {
		case requestStateParsingRequest:
			bytes, err = r.parseRequest(data)
			bytesParsed += bytes
			if err != nil {
				return bytesParsed, err
			}
		case requestStateParsingHeader:
			bytes, err = r.parseHeaders(data)
			bytesParsed += bytes
			if err != nil {
				return bytesParsed, err
			}
		case requestStateParsingBody:
			bytes, err = r.parseBody(data)
			bytesParsed += bytes
			if err != nil {
				return bytesParsed, err
			}
		}
		data = []byte("")

		if bytes == 0 {
			break
		}
		count += 1
		if count > 100 {
			break
		}
	}
	
	return bytesParsed, nil
}

func (r *Request) parseBody(data []byte ) (int, error) {
	r.RawRequest += string(data)
	r.Body = append(r.Body, r.RawRequest...)

	contentLength, err := r.Headers.Get("content-length")

	if err != nil || contentLength == ""{
		contentLength = "0"
	}
	r.RawRequest = ""
	bytesAdded := len(data)


	length, err := strconv.Atoi(contentLength)
	if err != nil {
		return 0, err
	}

	if len(r.Body) > length {
		r.RequestState = requestStateDone
		return bytesAdded, fmt.Errorf("Woah there cowboy. You've sent too much data")
	}

	if len(r.Body) == length && length != 0{
		r.RequestState = requestStateDone
	}

	return bytesAdded, nil
}

func FindCRLF(msg string) int {
	for i := 0; i < len(msg)-1; i++ {
		if msg[i] == '\r' && msg[i+1] == '\n' {
			return i 
		}
	}
	return -1
	
}
func parseRequestLine(msg string) ( int, *RequestLine, error) {
	endIndex := FindCRLF(msg)
	if endIndex == -1 {
		return 0, nil, nil
	}

	line := msg[0:endIndex]
	parts := strings.Split(line, " ")
	if (len(parts) != 3) {
		return  -1, nil, errors.New("invalid request line format. Request must have 3 parts")
	}
	method := parts[0]
	target := parts[1]
	versionString := parts[2]

	versionParts := strings.Split(versionString, "/")

	err := validateRequest(method, target, versionParts)
	if err != nil {
		return  -1, nil, err
	}

	return endIndex + 2, &RequestLine{
		HttpVersion: versionParts[1],
		RequestTarget: target,
		Method: method, }, nil
}

func (r *Request) parseRequest(data []byte) (int, error) {
	r.RawRequest += string(data)
	bytes, rl, err := parseRequestLine(r.RawRequest)
	if err != nil {
		r.RequestState = requestStateError
		return -1, err
	}
	if bytes == 0 {
		return bytes, nil
	}
	r.RequestLine = *rl
	r.RequestState = requestStateParsingHeader 
	r.RawRequest = r.RawRequest[bytes:]
	return bytes, nil
}

func (r *Request) parseHeaders(data []byte) (int, error) {
	r.RawRequest += string(data)
	bytes, done, err := r.Headers.Parse([]byte(r.RawRequest))
	if err != nil {
		r.RequestState = requestStateError
		return bytes, err
	}
	if done {
		r.RequestState = requestStateParsingBody
	}

	r.RawRequest = r.RawRequest[bytes:]

	return bytes, nil
}


func validateRequest(method, target string, versionParts []string) error{
	if target[0] != '/' {
		return errors.New("invalid request target: must start with '/'")
	}

	if method == "" {
		return errors.New("no method found in request")
	}
	if target == "" {
		return errors.New("no method found in request")
	}

	for _, c := range(method) {
		if !unicode.IsUpper(c) {
			return errors.New("invalid method: must contain only uppercase letters")
		}
	}

	if len(versionParts) != 2 || versionParts[0] != "HTTP" {
		return errors.New("invalid HTTP version format: only accept HTTP/1.1")
	}

	version := versionParts[1]
	if version != "1.1"{
		return errors.New("unexpected version: only accepting version 1.1")
	}

	return nil
}
