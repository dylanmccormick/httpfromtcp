package main

import (
	"bufio"
	"fmt"
	"log"
	"net"

	"github.com/dylanmccormick/httpfromtcp/internal/request"
)

func main(){
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("A listener error occurred: %s", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("A listener error occurred: %s", err)
		}
		log.Printf("Accepted a connection")

		scanner := bufio.NewReader(conn)


		rq, err := request.RequestFromReader(scanner)
		if err != nil {
			fmt.Errorf("Error occurred: %v", err)
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", rq.RequestLine.Method)
		fmt.Printf("- Target: %s\n", rq.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", rq.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		for k, v := range(rq.Headers){
			fmt.Printf("- %s: %s\n", k, v)
		}

	}



}


