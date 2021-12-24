package main

import (
	"GoChat/src/chat"
	"log"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:9000")
	if err != nil {
		log.Fatal(err)
	}
	go chat.Broadcaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go chat.HandleConn(conn)
	}
}
