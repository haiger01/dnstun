

const {
    DNS_Client  int = 0
    DNS_Server  int = 1
}

type struct DNSUtils {

    kind    int

    conn    *net.UDPConn

    top     string
    laddr   *UDPAddr
    ldns    *UDPAddr
}

func NewDNSClient(laddrstr, ldnsstr, topDomain string) (*DNSUtils, error){

    d := new(DNSUtils)
    d.kind = DNS_Client
    d.top = topDomain

    d.ldns, err = net.ResolveUDPAddr("udp", ldnsstr)
    if err != nil {
        return nil, err
    }

    d.laddr, err = net.ResolveUDPAddr("udp", laddrstr)
    if err != nil {
        return nil, err
    }

    /* Listen on UDP address laddr */
    d.conn, err = net.ListenUDP("udp", d.laddr)
    if err != nil {
        return nil, err
    }
}

func NewDNSServer(laddr, topDomain string) (*DNSUtils, error){

    d := new(DNSUtils)
    d.kind = DNS_Server
    d.top = topDomain

    d.laddr, err = net.ResolveUDPAddr("udp", laddrstr)
    if err != nil {
        return nil, err
    }
    d.ldns = d.laddr

    /* Listen on UDP address laddr */
    d.conn, err = net.ListenUDP("udp", d.laddr)
    if err != nil {
        return nil, err
    }
}

func (d *DNSUtils) Send(p []byte) error{
    if d.kind != DNS_Client {
        return fmt.Errorln("Send: Only used by Client\n")
    }
    _, err :=  d.conn.WriteToUDP(p, d.ldns)
    return err
}

func (d *DNSUtils) SendTo(p []byte, addr *UDPAddr) error{

    _, err := d.conn.WriteToUDP(p, addr)
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
