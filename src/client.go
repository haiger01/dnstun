package main

import (
    "log"
    "os"
    "fmt"
    "../lib/songgao/water"
    //"../lib/songgao/water/waterutil"
    "../lib/tonnerre/golang-dns"
)

type Client struct {
    ClientVAddr     *net.IPAddr
    ServerVAddr     *net.IPAddr

    //Laddr           string  // local address
    //LocalServer     string  // local dns server
    //TopDomain       string

    //DNSConn         *net.UDPConn

    DNS             *DNSUtils
    TUN             *Tunnel

    Buffer          map[int][]byte

    Running         bool
}

func NewClient(topdomain string, ldns string, laddr string,
                tunname string) (*Client, error) {

    c := new(Client)
    //c.TopDomain = topdomain
    //c.LocalServer = ldns

    c.Running = false

    c.ClientVAddr = nil
    c.ServerVAddr = nil

    // TODO
    c.DNS, err = NewDNS(laddr, ldns, topdomain)
    if err != nil {
        return nil, err
    }
    c.TUN, err = NewTunnel(tunnel)
    if err != nil {
        return nil, err
    }

    c.Buffer = make(map[int][]byte)
}

func (c *Client) Connect() error {

    // send connect request
    tunCmd = new(TUNPacket)
    tunCmd.Cmd = TUN_CMD_CREATE

    // TODO
    dnsPacket, err := c.DNS.Inject(tunCmd)
    c.DNS.Send(dnsPacket)

    go c.DNSRecv()
}

/*
func (c *Client) DNSSend(pkt []byte) error{
    _, err :=  DNSConn.Write(pkt)
    return err
}*/

func (c *Client) TUNRecv(){

    b = make([]byte, DEF_BUF_SIZE )
    for c.Running == true {

        n, err := c.TUN.Read(b)
        if err != nil {
            Error.Println(err)
            continue
        }

        s.DNS.InjectAndSendIPPacket(b[:n])
    }
}

func (c *Client) DNSRecv(){

    b := make([]byte, DEF_BUF_SIZE)
    for {
        n, addr, err := c.DNSConn.ReadFrom(b)
        if err != nil{
            Error.Println(err)
        }

        tun := c.tunnel.Unmarshal(b)
        switch tun.Cmd {
        case TUN_CMD_RESPONSE:
            cmd := tun.ToCmdPacket()
            c.ServerVAddr, err = net.ResolveIPAddr("ip", cmd.Server)
            if err != nil {
                Error.Println(err)
                continue
            }
            c.ClientVAddr, err = net.ResolveIPAddr("ip", cmd.Client)
            if err != nil {
                Error.Println(err)
                continue
            }

            c.Running = true
            go c.TUNRecv()

        case TUN_CMD_DATA:
            if c.Running == true{
                c.tunnel.Save(c.Buffer, tun)
            }
        default:
            Error.Println("Invalid TUN Cmd")
        }
    }
}


