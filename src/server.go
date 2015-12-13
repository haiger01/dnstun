package main

import (
    "log"
    "os"
    "fmt"
    "../lib/songgao/water"
    //"../lib/songgao/water/waterutil"
    "../lib/tonnerre/golang-dns"
)

type Conn struct {

    VAddr   *IPAddr
    PAddr   *UDPAddr

    buffer  map[int][]byte
    TUN     *Tunnel
}

type Server struct {

    LAddr   *UDPAddr
    VAddr   *IPAddr

    Routes  map[string]*Conn

    DNS     *DNSUtils
    //DNSConn *UDPConn

    TUN  *Tunnel
}

func NewServer(laddr string, tunname string, vaddr string) (*Server, error){

    s := new(Server)
    s.LAddr, err := net.ResolveUDPAddr(laddr)
    if err != nil {
        return nil, err
    }

    s.VAddr, err := net.ResolveIPAddr(vaddr)
    if err != nil {
        return nil, err
    }

    // TODO
    s.DNS, err = NewDNS(laddr, ldns, topdomain)
    if err != nil {
        return nil, err
    }

    s.TUN, err = NewTunnel(tunname)
    if err != nil {
        return nil, err
    }

}

func NewConn(vaddr *IPAddr, paddr *UDPAddr, tun *Tunnel) (*Conn){
    c := new(Conn)
    c.VAddr = vaddr
    c.PAddr = UDPAddr
    c.TUN = tun
    c.buffer = make(map[int][]byte)
    return c
}


func (c *Conn) Recv(t *TUNIPPacket){
    c.TUN.Save(c.buffer, t)
}

func (s *Server) AcquireVAddr() *IPAddr{

}

func (s *Server) DNSSend(p []byte) error{
    _, err :=  DNSConn.Write(pkt)
    return err
}

func (s *Server) DNSRecv(){

    b := make([]byte, DEF_BUF_SIZE)
    for {
        n, rpaddr, err := s.DNS.conn.ReadFromUDP(b)
        if err != nil{
            Error.Println(err)
        }

        /*
        conn, ok := s.Routes[addr.String()]
        if ok != true {
            Error.Printf("IP Packet %d not found\n", t.Id)
            continue
        }*/

        dnsPacket, err := s.DNS.Unmarshal(b[:n]) // TODO
        if err != nil {
            Error.Println(err)
            continue
        }

        tunPacket, err := s.TUN.Unmarshal(dnsPacket.Name) // TODO
        if err != nil {
            Error.Println(err)
            continue
        }

        switch tunPacket.GetCmd() {
        case TUN_CMD_CONNECT:

            rvaddr := s.AcquireVAddr()  //TODO
            lvaddr := s.VAddr

            // create new connection for the client
            conn := NewConn(rvaddr, rpaddr, s.TUN)
            s.Routes[rvaddr.String()] = conn

            t := new(TUNResponsePacket)
            t.Cmd = TUN_CMD_RESPONSE
            t.LAddr = lvaddr    // server's virtual address
            t.RAddr = rvaddr    // client's virtual address

            dnsPacket, err := dnsutils.Inject(t) // TODO
            s.DNS.Send(dnsPacket)

            Debug.Printf("Connected with %s\n", addr.String())

        case TUN_CMD_DATA:

            conn, err := s.FindByPAddr(rpaddr)  //TODO
            if err != nil{
                Debug.Println(err)
                continue
            }

            // cast packet to TUNIPPacket TODO: test if it works
            t, ok := tunPacket.(*TUNIPPacket)
            if ok != nil {
                Error.Printf("Unexpected cast fail from TUNPacket to TUNIPPacket\n")
                continue
            }else{
                conn.Recv(t)
            }

        case TUN_CMD_KILL:

            conn, err := s.FindByPAddr(rpaddr)
            if err != nil{
                Debug.Println(err)
                continue
            }
            delete(s.Routes_By_PAddr, conn.PAddr.String())
            delete(s.Routes_By_VAddr, conn.VAddr.String())
            Debug.Printf("Close Conn with %s\n", conn.VAddr.String())

        default:
            Error.Println("Invalid TUN Cmd")
        }
    }
}

func (s *Server) TUNRecv(){
    b = make([]byte, DEF_BUF_SIZE )
    for s.Running == true {

        n, err := s.TUN.Read(b)
        if err != nil {
            Error.Println(err)
            continue
        }

        s.DNS.InjectAndSendIPPacket(b[:n])
    }
}
