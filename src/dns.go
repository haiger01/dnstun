


type struct DNSUtils {

    laddr   *UDPAddr
    conn    *net.UDPConn
}

func NewDNS(laddrstr string) (*DNSUtils, error){

    d := new(DNSUtils)

    d.laddr, err = net.ResolveUDPAddr("udp4", laddrstr)
    if err != nil {
        return nil, err
    }

    conn, err = net.ListenUDP("udp4", d.laddr)
    if err != nil {
        return nil, err
    }
}


func (d *DNSUtils) Send(p []byte) error{
    _, err :=  d.conn.Write(pkt)
    return err
}

func (d *DNSUtils) Inject(tun *TUNPacket) ([]byte, int, error){

    // if in client side: inject into query 

    // if in server side inject into TXT
}

func (d *DNSUtils) Retreive([]byte) {

}

func (d *DNSUtils) Pack(*DNSPacket) []byte{

}

func (d *DNSUtils) Unpack(b []byte) *DNSPacket{

}

/* inject ip packet */
func (d *DNSUtils) InjectAndSendIPPacket(b []byte){

    ipPacket, err := ip.Unmarshal(b[:n]) // TODO
    if err != nil {
        Error.Println(err)
        continue
    }

    written := 0
    for written < n {

        // inject a ip fragment to DNS Packet
        // as much as possible

        // TODO
        dnsPacket, w, err := c.DNS.Inject(ipPacket, b[written:])
        s.DNS.Send(dnsPacket)
        if err != nil {
            Error.Println(err)
            continue
        }
        written += w
    }
}
