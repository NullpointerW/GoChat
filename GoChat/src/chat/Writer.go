package chat

import (
	"fmt"
	"net"
)

const title = "\n\n\t\t\tThe Go Chat \n\n\t\t\t Ver 0.1.2\n"

func clientWriter(conn net.Conn, ch <-chan string, prCh <-chan string, prMsg <-chan client, WprMsg chan<- client, errMsg <-chan string) {
	fmt.Fprintln(conn, title)
	for {
		select {
		case err := <-errMsg:
			fmt.Fprintln(conn, err)
		case cli := <-prMsg:
			fmt.Fprintln(conn, prefix+cli.addr+" want to establish a private chat with you \n Y/N?\n")
			WprMsg <- cli
		case msg := <-ch:
			fmt.Fprintln(conn, msg)
		case msg := <-prCh:
			fmt.Fprintln(conn, msg)
			for msg = range prCh {
				fmt.Fprintln(conn, msg)
			}
		}
	}
}
