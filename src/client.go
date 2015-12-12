package main

import (
    "log"
    "os"
    "fmt"
    "../lib/songgao/water"
    //"../lib/songgao/water/waterutil"
    "../lib/tonnerre/golang-dns"
)

var (
    Debug *log.Logger = log.New(os.Stderr, "Debug: ", log.Lshortfile)
    Error *log.Logger = log.New(os.Stderr, "Error: ", log.Lshortfile)
)

func OpenTUN() (*water.Interface){
    ifce, err := water.NewTUN("tun66")
    if err != nil {
        panic(err)
    }
    return ifce
}

func OpenDNSClient() (*dns.Client){

    c := new(dns.Client)
    c.Net = "udp"

    return c
}

func Recv(){

    for{
        // reading from DNS socket

        // decapsulate ip packet from dns

        // print
    }
}

func Send(){
    for{
        // receive ip packet from TUN
    }
}

func NewDNSPacket(raw []byte) *dns.Msg {

    msg := new(dns.Msg)
    //msg.SetQuestion("www.google.com.", dns.TypeA)
    msg.SetQuestion("b.jannotti.com.", dns.TypeA)
    msg.RecursionDesired =true
    txt := new(dns.TXT)
    txt.Txt = []string{string(raw)}

    //msg.Insert([]dns.RR{txt})

    Debug.Println(msg.String())
    return msg
}

func main(){

    // open TUN connection
    ifce := OpenTUN();
    client := OpenDNSClient();

    buffer := make([]byte, 1600)
    for {
        n, err := ifce.Read(buffer)
        if err != nil {
            Error.Println(err)
        }
        fmt.Printf("Read %d bytes\n", n)

        // encapsulate into dns packet
        msg := NewDNSPacket(buffer[:n])

        // send out dns TODO: local dns server
        res, rtt, err := client.Exchange(msg, "172.31.0.2:53")
        if err != nil {
            Error.Println(err)
        }

        Debug.Printf("DNS response, rtt: %s : %s\n", rtt.String(), res.String())
    }
}



