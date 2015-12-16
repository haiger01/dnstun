package main

import (
	"../lib/tun"
	"flag"
	"fmt"
	"strings"
    "bufio"
    "os"
)

func rpl(s *tun.Server) {

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Printf("server > ")
		cmd := strings.Split(scanner.Text(), " ")
		switch cmd[0] {
		case "ping":
			fmt.Println("server ping client not implemented")
		case "info":
			s.Info()
		case "quit", "exit":
			fmt.Println("Goodbye!")
			return
		}
	}
	if err := scanner.Err(); err != nil {
		Error.Printf("reading standard input: %s\n", err)
	}
}

func main() {

	topDomainPtr := flag.String("d", DEF_TOP_DOMAIN, "Top Domain")
	laddrPtr := flag.String("l", DEF_DOMAIN_PORT, "Address of DNS Server")
	vaddrPtr := flag.String("v", "192.168.3.1", "Virtual IP Address of Tunneling Server")
	tunPtr := flag.String("t", "tun66", "Name of TUN Interface")

	flag.Parse()

	server, err := tun.NewServer(*topDomainPtr,
		*laddrPtr,
		*vaddrPtr,
		*tunPtr)
	if err != nil {
		Error.Println(err)
		return
	}

	go server.DNSRecv()
	go server.TUNRecv()
	rpl(server)

	return

}
