package main

import (
	"../lib/tun"
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

func testPing() {

}

func rpl(c *tun.Client) {

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		//fmt.Println(scanner.Text()) // Println will add back the final '\n'
		fmt.Printf("client> ")
		cmd := strings.Split(scanner.Text(), " ")
		switch cmd[0] {
		case "ping":
			testPing()
		case "info":
			c.Info()
		case "kill":
			fmt.Printf("kill not implemented\n")
		case "quit", "exit":
			fmt.Printf("Goodbye!\n")
			return
		}
	}
	if err := scanner.Err(); err != nil {
		Error.Printf("reading standard input: %s\n", err)
	}
}

func main() {

	topDomainPtr := flag.String("d", DEF_TOP_DOMAIN, "Top Domain")
	ldnsPtr := flag.String("n", DEF_LOCAL_DNS, "Address of Local DNS Server")
	laddrPtr := flag.String("l", ":4000", "Addrss of DNS Client")
	tunPtr := flag.String("t", "tun66", "Name of TUN Interface")

	flag.Parse()

	client, err := tun.NewClient(*topDomainPtr,
		*ldnsPtr,
		*laddrPtr,
		*tunPtr)
	if err != nil {
		Error.Println(err)
		return
	}

	err = client.Connect()
	if err != nil {
		Error.Println(err)
	}

	rpl(client)

	return
}
