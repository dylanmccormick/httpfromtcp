package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"bufio"
)

func main(){
	udpAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatalf("Error resolving UDP address: %v", err)
	}
	udpConn, err:= net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatalf("Error Dialing UDP address: %v", err)
	}
	defer udpConn.Close()

	for {
		fmt.Print(">")
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Unable to read with error: %v", err)
		}
		fmt.Printf("You entered: %s", line)
		b := []byte(line)
		udpConn.Write(b)
	}

}
