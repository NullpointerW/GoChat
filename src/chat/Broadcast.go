package chat

import "log"

func Broadcaster() {
	clients := make(map[client]bool) // all connected clients
	for {
		select {
		case addrs := <-listReq:
			log.Printf("from slice ptr:%p ", addrs)
			for c, _ := range clients {
				addrs = append(addrs, c.addr)
				log.Printf("chan slice ptr:%p ", addrs)
			}
			listReq <- addrs
		case sch := <-searchReq:
			var find bool
			for cli := range clients {
				if sch.addr == cli.addr {
					sch.tar.searchRes <- cli
					find = true
					break
				}
			}
			if !find {
				sch.tar.searchRes <- client{}
			}
		case msg := <-messages:
			// Broadcast incoming message to all
			// clients' outgoing message channels.
			for cli := range clients {
				cli.prMu.Lock()
				if !*cli.prState {
					cli.ch <- msg
				}
				cli.prMu.Unlock()
			}

		case cli := <-entering:
			clients[cli] = true

		case cli := <-leaving:
			delete(clients, cli)
			close(cli.prMsg)
			close(cli.prCh)
			close(cli.ch)
		}
	}
}
