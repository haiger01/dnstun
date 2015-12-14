package tun

import (
    "net"
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

    var err error
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
    tunPacket := new(TUNCmdPacket)
    tunPacket.Cmd = TUN_CMD_CONNECT

    // Inject the TUN Packet to a DNS Packet
    msgs, err := c.DNS.Inject(tunPacket)
    if err != nil {
        return err
    }

    // Listening on the port, wating for incoming DNS Packet
    go c.DNSRecv()

    // Send DNS Packet to Local DNS Server
    for i:= 0; i<len(msgs); i++{
        packet, err := msgs[i].Pack()
        if err != nil {
            return err
        }
        err = c.DNS.Send(packet)
        if err != nil {
            return err
        }
    }
    return nil
}

func (c *Client) DNSRecv(){

    b := make([]byte, DEF_BUF_SIZE)
    for {
        // rpaddr : the public UDP Addr of remote DNS Server
        n, _, err := c.DNS.Conn.ReadFrom(b)
        if err != nil{
            Error.Println(err)
        }

        dnsPacket := new(dns.Msg)
        err = dnsPacket.Unpack(b[:n])
        if err != nil {
            Error.Println(err)
            continue
        }
        tunPacket, err := c.DNS.Retrieve(dnsPacket) // TODO
        if err != nil {
            Error.Println(err)
            continue
        }

        switch tunPacket.GetCmd() {
        case TUN_CMD_RESPONSE:

            res, ok := tunPacket.(*TUNResponsePacket)
            if !ok {
                Error.Println("Fail to Convert TUN Packet\n")
                continue
            }

            c.ServerVAddr  = new(net.IPAddr)
            c.ClientVAddr  = new(net.IPAddr)
            *c.ServerVAddr = *res.Server
            *c.ClientVAddr = *res.Client

            /*
            *(c.ServerVAddr)("ip", res.Server)
            if err != nil {
                Error.Println(err)
                continue
            }
            c.ClientVAddr, err = net.ResolveIPAddr("ip", res.Client)
            if err != nil {
                Error.Println(err)
                continue
            }*/

            c.Running = true
            go c.TUNRecv()

        case TUN_CMD_DATA:

            if c.Running == true{

                t, ok := tunPacket.(*TUNIPPacket)
                if !ok {
                    Error.Println("Fail to Convert TUN Packet\n")
                    continue
                }
                c.TUN.Save(c.Buffer, t)
            }
        default:
            Error.Println("Invalid TUN Cmd")
        }
    }
}

func (c *Client) TUNRecv(){

    b := make([]byte, DEF_BUF_SIZE )
    for c.Running == true {

        n, err := c.TUN.Read(b)
        if err != nil {
            Error.Println(err)
            continue
        }

        err = c.DNS.InjectAndSendTo(b[:n], c.DNS.LDns)
        if err != nil {
            Error.Println(err)
            continue
        }
    }
}
