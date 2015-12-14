package main

import (
    "flag"
    "../lib/tun"
)

func main(){

    topDomainPtr:= flag.String("d", DEF_TOP_DOMAIN, "Top Domain")
    laddrPtr    := flag.String("l", DEF_DOMAIN_PORT, "Address of DNS Server")
    vaddrPtr    := flag.String("v", "192.168.3.1", "Virtual IP Address of Tunneling Server")
    tunPtr      := flag.String("t", "tun66", "Name of TUN Interface")

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
    server.TUNRecv()

}
