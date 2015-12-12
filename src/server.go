package main

import (
    "fmt"
    "log"
    "os"
    "net"
    "io"
    //"fmt"
    //"../lib/songgao/water"
    //"../lib/songgao/water/waterutil"
    "../lib/tonnerre/golang-dns"
)

var (
    Debug *log.Logger = log.New(os.Stderr, "Debug: ", log.Lshortfile)
    Error *log.Logger = log.New(os.Stderr, "Error: ", log.Lshortfile)
)

func udp(){

    addr, err := net.ResolveUDPAddr("udp","0.0.0.0:53")
    if err != nil {
        Error.Println(err)
        return
    }


    b := make([]byte, 1500)
    conn, _ := net.ListenUDP("udp4", addr)
    for{
        n, _ := conn.Read(b)

        if n > 0{
            msg := new(dns.Msg)
            msg.Unpack(b[:n])

            fmt.Printf("-------------------------------------\n")
            fmt.Printf("udp: received %d, %s\n",n, msg.String())
        }
    }
}

func main(){
    /*
    server := new(dns.Server)
    server.Net = "tcp"
    server.ListenAndServe()
    */

    go udp()

    listener, err := net.Listen( "tcp", ":53")
    if err != nil {
        Error.Println(err)
        return
    }

    defer listener.Close()

    for {
        conn, err := listener.Accept()
        if( err != nil){
            Error.Println(err)
        }
        go func( conn net.Conn){
            b := make([]byte,1500)
            for{
                n, e := conn.Read(b)
                fmt.Println(string(b[:n]))
                if e == io.EOF {
                    conn.Close()
                    return
                }
            }
            conn.Close()
        }(conn)
    }
}

