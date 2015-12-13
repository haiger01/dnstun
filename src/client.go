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

const (
    DEF_BUF_SIZE int = 1500
)

type Client struct {
    ClientVAddr     *IPAddr
    ServerVAddr     *IPAddr

    Laddr           string  // local address
    LocalServer     string  // local dns server
    TopDomain       string

    DNSConn         *net.UDPConn
    TUNConn         *water.Interface

    Buffer          map[int][]byte

    Running         bool
}

func NewClient(topdomain string, ldns string) error {

    c := new(Client)
    c.TopDomain = topdomain
    c.Laddr = ...   // TODO
    c.LocalServer = ldns

    c.Running = false

    c.ClientVAddr = nil
    c.ServerVAddr = nil
    c.DNSConn = nil
    c.TUNConn = nil

    c.Buffer = make(map[int][]byte)
}

func (c *Client) Connect() error {

    // connect with UDP
    DNSConn, err = net.ListenUDP("udp4", c.Laddr) // TODO
    if err != nil {
        Error.Printl
    }

    // send connect request
    tunCmd = new(TUNPacket)
    tunCmd.Cmd = TUN_CMD_CREATE
    dnsPacket, err := c.dns.InjectTunCmd(tunCmd)
    c.DNSSend(dnsPacket)

    go c.DNSRecv()
}

func (c *Client) DNSSend(pkt []byte) error{

    _, err :=  DNSConn.Write(pkt)
    return err
}

func (c *Client) TUNRecv(){

    b = make([]byte, DEF_BUF_SIZE )
    for c.Running == true {

        n, err := TUNConn.Read(b)
        if err != nil {
            Error.Println(err)
            continue
        }

        ipp, err := ip.Unmarshal(b[:n])
        if err != nil {
            Error.Println(err)
            continue
        }

        written := 0
        for written < n {
            // encapsulate tunp to DNS Packet
            dnsPacket, w, err := c.dns.Encapsulate(ipp, b, written)
            c.DNSSend(dnsPacket)
            if err != nil {
                Error.Println(err)
                continue
            }

            written += w
        }
    }
}

func (c *Client) Upload(pkt []byte) error{

    n, err := c.TUNConn.Write(pkt)
    if err != nil {
        Error.Println(err)
        return err
    }
    if n != len(pkt){
        return fmt.Errorf("Short write %d, should be %d", n, len(pkt))
    }
}

func (c *Client) SaveTUNPacket(tun *TUNPacket){

    if tun.Offset == 0 && tun.More == false {
        pkt := tun.Payload
        c.Upload(pkt)   // send to upper layer
        return
    }
    pkt, ok := c.Buffer[tun.Id]
    if ok {
        if tun.Offset == len(pkt) {
            pkt := append(pkt, tun.Payload)
            if tun.More == false{
                c.Upload(pkt)
                delete(c.Buffer, tun.Id)
            }else{
                c.Buffer[tun.Id] = pkt
            }
        }
    }else{
        c.Buffer[tun.Id] = tun.Payload
    }
}

func (c *Client) DNSRecv(){

    b := make([]byte, DEF_BUF_SIZE)
    for {
        n, addr, err := c.DNSConn.ReadFrom(b)
        if err != nil{
            Error.Println(err)
        }

        tun := tunnel.Unmarshal(b)
        switch tun.Cmd {
        case TUN_CMD_RESPONSE:
            cmd := tun.ToCmdPacket()
            c.ServerVAddr = cmd.Server
            c.ClientVAddr = cmd.Client
            c.Running = true
            go c.TUNRecv()

        case TUN_CMD_DATA:
            if c.Running == true{
                c.SaveTUNPacket(tun)
            }
        default:
            Error.Println("Invalid TUN Cmd")
        }
    }
}


