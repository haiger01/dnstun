package tun

import (
    "log"
    "os"
    "fmt"
    "..//ip"
    "../songgao/water"
    //"../songgao/water/waterutil"
    "../tonnerre/golang-dns"
)

type Client struct {
    ClientVAddr     *net.IPAddr
    ServerVAddr     *net.IPAddr

    DNS             *DNSUtils
    TUN             *Tunnel

    Buffer          map[int][]byte

    Running         bool
}

func NewClient(topDomain, ldns, laddr, tunName string) (*Client, error) {

    c := new(Client)
    c.Running = false

    /* Will be filled after connected with server */
    c.ClientVAddr = nil
    c.ServerVAddr = nil

    c.DNS, err = NewDNSClient(laddr, ldns, topDomain)
    if err != nil {
        return nil, err
    }
    c.TUN, err = NewTunnel(tunName)
    if err != nil {
        return nil, err
    }

    c.Buffer = make(map[int][]byte)
    return c, nil
}

func (c *Client) Connect() error {

    // Create a TUN Packet
    tunPacket = new(TUNCmdPacket)
    tunPacket.Cmd = TUN_CMD_CREATE

    // Inject the TUN Packet to a DNS Packet
    dnsPacket, err := c.DNS.Inject(tunPacket)
    if err != nil {
        return err
    }

    // Listening on the port, wating for incoming DNS Packet
    go c.DNSRecv()

    // Send DNS Packet to Local DNS Server
    err := c.DNS.Send(dnsPacket)
    if err != nil {
        return err
    }
    return nil
}

func (c *Client) DNSRecv(){

    b := make([]byte, DEF_BUF_SIZE)
    for {
        // rpaddr : the public UDP Addr of remote DNS Server
        n, rpaddr, err := c.DNSConn.ReadFrom(b)
        if err != nil{
            Error.Println(err)
        }

        dnsPacket, err := s.DNS.Unpack(b[:n]) // TODO
        if err != nil {
            Error.Println(err)
            continue
        }
        tunPacket, err := s.DNS.Retrieve(dnsPacket) // TODO
        if err != nil {
            Error.Println(err)
            continue
        }

        switch tunPacket.GetCmd() {
        case TUN_CMD_RESPONSE:

            res, ok := tunPacket.(*TUNResponsePacket)
            if !ok {
                Error.Println("Fail to Convert TUN Packet\n"
                panic()
                continue
            }

            c.ServerVAddr, err = net.ResolveIPAddr("ip", res.Server)
            if err != nil {
                Error.Println(err)
                continue
            }
            c.ClientVAddr, err = net.ResolveIPAddr("ip", res.Client)
            if err != nil {
                Error.Println(err)
                continue
            }

            c.Running = true
            go c.TUNRecv()

        case TUN_CMD_DATA:

            if c.Running == true{

                t, ok := tunPacket.(*TUNResponsePacket)
                if !ok {
                    Error.Println("Fail to Convert TUN Packet\n"
                    panic()
                    continue
                }
                c.tunnel.Save(c.Buffer, t)
            }
        default:
            Error.Println("Invalid TUN Cmd")
        }
    }
}

func (c *Client) TUNRecv(){

    b = make([]byte, DEF_BUF_SIZE )
    for c.Running == true {

        n, err := c.TUN.Read(b)
        if err != nil {
            Error.Println(err)
            continue
        }

        err := c.DNS.InjectAndSendTo(b[:n], c.DNS.LDns)
        if err != nil {
            Error.Println(err)
            continue
        }
    }
}
