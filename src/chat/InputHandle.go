package chat

import (
	"bufio"
	"log"
	"net"
	"strings"
	"sync"
)

const (
	prefix      = "[Go chat] "
	helpContent = "\n\n\t\t\tThe Go Chat \n\n\t\t\t Ver 0.1.2\n\nCommands list :\n\n\t-P [IP:PORT]\t\t" +
		"try to establish a private chat with a specify addr \n\t-H\t\t\tshow all commands list \n\t" +
		"-Q\t\t\tclose current private chat \n\t-C\t\t\tcLear screen \n\t-L\t\t\tshow all connect clients addr\n"
	cls  = "\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n"
	tail = "\ngo chat:"
)

type search struct {
	addr string
	tar  client
}

var (
	entering  = make(chan client)
	leaving   = make(chan client)
	messages  = make(chan string) // all incoming client messages
	searchReq = make(chan search)
	listReq   = make(chan []string)
)

type client struct {
	ch        chan<- string // an outgoing message channel
	prCh      chan<- string
	addr      string
	prMsg     chan<- client
	etlFlag   chan<- bool
	searchRes chan<- client
	prState   *bool
	prMu      *sync.Mutex
	errMsg    chan<- string
}

func HandleConn(conn net.Conn) {
	prMode := false
	who := conn.RemoteAddr().String()
	ch := make(chan string)   // outgoing client messages
	prCh := make(chan string) // outgoing client  private messages
	prMsg := make(chan client)
	WprMsg := make(chan client, 1)
	etlFlag := make(chan bool)
	searchRes := make(chan client)
	errMsg := make(chan string)
	prState := make(chan bool)
	prChange := make(chan bool, 1)
	cli := client{ch, prCh, who, prMsg, etlFlag, searchRes, new(bool), new(sync.Mutex), errMsg}
	log.Printf(" clinet:%v", cli)
	go clientWriter(conn, ch, prCh, prMsg, WprMsg, errMsg, prState, prChange)
	log.Printf("%s connected", who)
	ch <- prefix + "You are " + who
	messages <- prefix + who + " has arrived"
	entering <- cli
	var PCli client
	input := bufio.NewScanner(conn)
	for input.Scan() {
		text := input.Text()
		select {
		case st := <-prChange:
			prMode = st
		default:

		}
		if text != "" {
			fmtStr := strings.Replace(text, " ", "", -1)
			fmtStr = strings.ToUpper(fmtStr)
			r := fmtStr[0]
			if r == '-' {
				arg := fmtStr[1]
				switch arg {
				case 'P':
					if prMode {
						errMsg <- "[ERROR] -" + string(arg) + ": you are already in a private chat"
						continue
					}
					text = strings.ToUpper(text)
					text = strings.Replace(text, " ", "", -1)
					text = strings.Replace(text, "-P", "", -1)
					req := search{tar: cli, addr: text}
					searchReq <- req
					PCli = <-searchRes
					if PCli.addr == "" {
						errMsg <- prefix + "[ERROR] -p :can not find address"
						continue
					}
					log.Printf(" Point clinet:%v", PCli)
					PCli.prMsg <- cli
					if flag := <-etlFlag; flag {
						prState <- true
						prMode = true
						cli.prMu.Lock()
						*cli.prState = true
						cli.prMu.Unlock()
						errMsg <- prefix + " " + "a private chat established with " + PCli.addr
					} else {
						errMsg <- prefix + PCli.addr + " " + "reject your establish request"
					}

				case 'H':
					cli.ch <- helpContent
				case 'L':
					var list []string
					log.Printf("slice ptr:%p ", list)
					listReq <- list
					list = <-listReq
					respList := "\nAddress List :\n\n"
					for _, e := range list {
						respList += "\t" + e + "\n"
					}
					cli.ch <- respList
				case 'C':
					msg := cls
					cli.ch <- msg
				case 'Q':
					prMode = false
					prState <- false
					PCli.errMsg <- "exit" + cli.addr
					cli.prMu.Lock()
					*cli.prState = false
					cli.prMu.Unlock()
					PCli.prMu.Lock()
					*PCli.prState = false
					PCli.prMu.Unlock()
				default:
					cli.ch <- prefix + "Unknown commands '" + string(arg) + "' use -H to show more information"
				}
				continue
			}
		}
		if prMode { //-p [HOST:PORT]
			PCli.prCh <- prefix + "[private]" + who + ": " + text
			continue
		}
		select {
		case PCli = <-WprMsg:
			if strings.ToUpper(text) == "Y" {
				prState <- true
				prMode = true
				PCli.etlFlag <- true
				cli.prMu.Lock()
				*cli.prState = true
				cli.prMu.Unlock()
				errMsg <- prefix + PCli.addr + " " + "a private chat established"
			} else {
				PCli.etlFlag <- false
			}
			continue
		default:
		}
		messages <- prefix + who + ": " + text
	}
	// NOTE: ignoring potential errors from input.Err()

	leaving <- cli
	messages <- prefix + who + " has left"
	conn.Close()
	log.Printf("%s closed connect", who)

}
