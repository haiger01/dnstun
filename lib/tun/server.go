package tun

import (
    "net"
    "../tonnerre/golang-dns"
    "../ip"
    "fmt"
)

type Conn struct {

    VAddr   *net.IPAddr
    PAddr   *net.UDPAddr // 
    UserId    int

    InChan   chan TUNPacket

    Buffer  map[int][]byte
    TUN     *Tunnel
    DNS     *DNSUtils
}

type Server struct {

    /* Physical Address DNS Server Listening on */
    //LAddr   *UDPAddr

    /* Virtual Address in TUN Virtual Interface */
    VAddr   *net.IPAddr

    Routes_By_VAddr  map[string]*Conn
    Routes_By_UserId   map[int]*Conn

    DNS     *DNSUtils
    TUN     *Tunnel
}

func NewServer(topDomain, laddr, vaddr, tunName string) (*Server, error){

    s := new(Server)
    /*
    s.LAddr, err := net.ResolveUDPAddr("udp", laddr)
    if err != nil {
        return nil, err
    }*/

    var err error
    s.VAddr, err = net.ResolveIPAddr("ip", vaddr)
    if err != nil {
        return nil, err
    }

    s.Routes_By_VAddr = make(map[string]*Conn)
    s.Routes_By_UserId = make(map[int]*Conn)

    s.DNS, err = NewDNSServer(laddr, topDomain)
    if err != nil {
        return nil, err
    }

    s.TUN, err = NewTunnel(tunName)
    if err != nil {
        return nil, err
    }
    return s, nil
}

func (s *Server) NewConn(vaddr *net.IPAddr, user int) (*Conn){
    c := new(Conn)
    c.VAddr = vaddr
    //c.PAddr = paddr

    c.InChan = make(chan TUNPacket, 200)

    c.TUN = s.TUN
    c.DNS = s.DNS
    c.Buffer = make(map[int][]byte)
    return c
}


func (c *Conn) Recv(tunPacket TUNPacket) error{

    // cast packet to TUNIPPacket TODO: test if it works
    t, ok := tunPacket.(*TUNIPPacket)
    if !ok {
        return fmt.Errorf("Unexpected cast fail from TUNPacket to TUNIPPacket\n")
    }
    err := c.TUN.Save(c.Buffer, t)
    if err != nil {
        return err
    }
    return nil
}

func (s *Server) NormalReply(msg *dns.Msg, paddr *net.UDPAddr) error{
    return nil
}

func (c *Conn) Reply(msg *dns.Msg, paddr *net.UDPAddr) error{

    //b := make([]byte, 1600)

    reply := new(dns.Msg)
    reply.SetReply(msg)

    select {
    case tunPacket := <-c.InChan :
    // TODO
    // There're pending TUN Packets, Inject it into DNS Reply Packet
    // And Send Back
        Error.Printf("Not Implemented, %s", tunPacket.GetUserId())

        c.DNS.Reply(msg, tunPacket, paddr)

    default:
    // TODO
    // there's no pending TUN Packet to be sent,
    // just reply the request

        // normal reply 
        t := &TUNCmdPacket{ TUN_CMD_ACK, c.UserId}
        return c.DNS.Reply(msg, t, paddr)

        /*
        domain := msg.Question[0].Name
        txt, err := dns.NewRR(domain + " 1 IN TXT abcdeabcde")
        reply.Answer = make([]dns.RR, 1)
        reply.Answer[0] = txt

        b, err := reply.Pack()
        if err != nil {
            Error.Println(err)
            return err
        }

        err = c.DNS.SendTo(paddr, b)
        if err != nil{
            Error.Println(err)
            return err
        }*/
    }
    return nil
}

func (s *Server) AcquireVAddr() *net.IPAddr{
    //addr := new(net.IPAddr)
    //*addr = *s.NextIPAddr
    // TODO

    addr, err := net.ResolveIPAddr("ip4", "192.168.3.4")
    if err != nil{
        Error.Println(err)
        return nil
    }
    return addr
}
func (s *Server) AcquireUserId()  int{

    // TODO

    return 124
}

func (s *Server) FindConnByVAddr(addr string) (*Conn, error){

    conn, ok := s.Routes_By_VAddr[addr]
    if !ok {
        return nil, fmt.Errorf("Cannot find Connection for Addr %s\n",
                    addr)
    }
    return conn, nil
}

func (s *Server) FindConnByUserId(user int) (*Conn, error){

    conn, ok := s.Routes_By_UserId[user]
    if !ok {
        return nil, fmt.Errorf("Cannot find Connection for UserId %d\n",
                    user)
    }
    return conn, nil
}

func (s *Server) DNSRecv(){

    b := make([]byte, DEF_BUF_SIZE)
    for {
        n, rpaddr, err := s.DNS.Conn.ReadFromUDP(b)
        if err != nil{
            Error.Println(err)
            continue
        }

        dnsPacket := new(dns.Msg)
        err = dnsPacket.Unpack(b[:n])
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
        case TUN_CMD_CONNECT:

            rvaddr := s.AcquireVAddr()  // TODO
            user := s.AcquireUserId()     // TODO

            // create new connection for the client
            conn := s.NewConn(rvaddr, user)
            s.Routes_By_VAddr[rvaddr.String()] = conn
            s.Routes_By_UserId[user] = conn

            // TODO: reply with response in TXT

            t := &TUNResponsePacket{TUN_CMD_RESPONSE,
                                    user,
                                    s.VAddr,
                                    rvaddr}

            err := s.DNS.Reply(dnsPacket, t, rpaddr)
            if err != nil{
                Error.Println(err)
                continue
            }

            /*
            msgs, err := s.DNS.Inject(t) // TODO
            if err != nil {
                Error.Println(err)
                continue
            }

            if len(msgs) != 1 {
                Error.Println("CONNECT: should be one DNS Packet\n")
                continue
            }

            binary, err := msgs[0].Pack()
            err = s.DNS.SendTo( conn.PAddr, binary)
            if err != nil {
                Error.Println(err)
                continue
            }*/

            Debug.Printf("Connected with %s\n", conn.PAddr.String())
            continue

        case TUN_CMD_KILL:

            conn, err := s.FindConnByUserId( tunPacket.GetUserId() )
            if err != nil {
                Error.Println(err)
                continue
            }

            delete(s.Routes_By_UserId, conn.UserId)
            delete(s.Routes_By_VAddr, conn.VAddr.String())
            // option: remove user from user pool
            // remove vaddr from vaddr pool
            Debug.Printf("Close Conn with %s\n", conn.VAddr.String())

            // normal reply 
            t := &TUNCmdPacket{TUN_CMD_ACK, conn.UserId}
            err = s.DNS.Reply(dnsPacket, t, rpaddr)
            if err != nil{
                Error.Println(err)
                continue
            }

        case TUN_CMD_DATA:

            conn, err := s.FindConnByUserId( tunPacket.GetUserId() )
            if err != nil {
                Error.Println(err)
                continue
            }

            err = conn.Recv(tunPacket)
            if err != nil{
                Error.Println(err)
                continue
            }

            // normal reply this message
            t := &TUNCmdPacket{TUN_CMD_ACK, conn.UserId}
            err = s.DNS.Reply(dnsPacket, t, rpaddr)
            if err != nil{
                Error.Println(err)
                continue
            }

        case TUN_CMD_EMPTY:
            // user.cmd.domain.com

            conn, err := s.FindConnByUserId( tunPacket.GetUserId() )
            if err != nil {
                Error.Println(err)
                continue
            }

            // reply 
            err = conn.Reply(dnsPacket, rpaddr)
            if err != nil{
                Error.Println(err)
            }

        case TUN_CMD_ACK:
            // xxx.domain.com

            // normal reply
            conn, err := s.FindConnByUserId( tunPacket.GetUserId() )
            if err != nil {
                fmt.Println("cannot find conn by user")
            }
            t := &TUNCmdPacket{TUN_CMD_ACK, conn.UserId}
            err = s.DNS.Reply(dnsPacket, t, rpaddr)
            if err != nil{
                Error.Println(err)
                continue
            }

        default:
            // Reply with normal DNS Response
            Error.Println("Invalid TUN Cmd -- Not Implemented")
        }
    }
}

func (s *Server) TUNRecv(){

    b := make([]byte, DEF_BUF_SIZE )
    for {

        n, err := s.TUN.Read(b)
        if err != nil {
            Error.Println(err)
            continue
        }

        ippkt := new(ip.IPPacket)
        err = ippkt.Unmarshal(b[:n])
        if err != nil {
            Error.Println(err)
            continue
        }
        Debug.Printf("TUN: IP Packet from %s to %s\n",
            ip.IPAddrInt2Str(ippkt.Header.Src),
            ip.IPAddrInt2Str(ippkt.Header.Dst))

        rvaddrStr := ip.IPAddrInt2Str(ippkt.Header.Dst)

        conn, err := s.FindConnByVAddr(rvaddrStr)
        if err != nil{
            Debug.Println(err)
            continue
        }

        tunPacket := new(TUNIPPacket)
        tunPacket.Cmd = TUN_CMD_DATA
        tunPacket.Id = int(ippkt.Header.Id)
        tunPacket.Offset = 0
        tunPacket.More = false
        tunPacket.Payload = b[:n]

        conn.InChan <- tunPacket

        /*
        msgs, err := s.DNS.Inject(tunPacket)
        if err != nil {
            Error.Println(err)
            continue
        }

        for i:=0; i<len(msgs); i++{
            conn.InChan <- msgs[i]
            continue
        }*/
    }
}
