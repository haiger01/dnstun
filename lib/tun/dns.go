package tun
import (
    "net"
    "fmt"
    "strings"
    "encoding/base32"
    "strconv"
    "../tonnerre/golang-dns"
    "../ip"
)

const (
    DNS_Client  int = 0
    DNS_Server  int = 1
)

type DNSUtils struct {

    Kind    int
    Conn    *net.UDPConn
    TopDomain     string
    LAddr   *net.UDPAddr
    LDns    *net.UDPAddr
}

func NewDNSClient(laddrstr, ldnsstr, topDomain string) (*DNSUtils, error){

    d := new(DNSUtils)
    d.Kind = DNS_Client
    d.TopDomain = topDomain

    var err error
    d.LDns, err = net.ResolveUDPAddr("udp", ldnsstr)
    if err != nil {
        return nil, err
    }

    d.LAddr, err = net.ResolveUDPAddr("udp", laddrstr)
    if err != nil {
        return nil, err
    }

    /* Listen on UDP address laddr */
    d.Conn, err = net.ListenUDP("udp", d.LAddr)
    if err != nil {
        return nil, err
    }
    return d, nil
}

func NewDNSServer(laddrstr, topDomain string) (*DNSUtils, error){

    d := new(DNSUtils)
    d.Kind = DNS_Server
    d.TopDomain = topDomain

    var err error
    d.LAddr, err = net.ResolveUDPAddr("udp", laddrstr)
    if err != nil {
        return nil, err
    }
    d.LDns = d.LAddr

    /* Listen on UDP address laddr */
    d.Conn, err = net.ListenUDP("udp", d.LAddr)
    if err != nil {
        return nil, err
    }
    return d, nil
}

func (d *DNSUtils) NewDNSPacket(t TUNPacket) (*dns.Msg, error){

    switch t.GetCmd(){
    case TUN_CMD_CONNECT:
        labels := []string{string(TUN_CMD_CONNECT), d.TopDomain}
        domain := strings.Join(labels, ".")

        msg := new(dns.Msg)
        msg.SetQuestion(domain, dns.TypeTXT)
        msg.RecursionDesired = true
        return msg, nil

    default:
        return nil, fmt.Errorf("NewDNSPacket: Invalid TUN CMD\n")
    }
}

func (d *DNSUtils) Send(p []byte) error{
    if d.Kind != DNS_Client {
        return fmt.Errorf("Send: Only used by Client\n")
    }
    _, err :=  d.Conn.WriteToUDP(p, d.LDns)
    return err
}

func (d *DNSUtils) SendTo(addr *net.UDPAddr, p []byte) error{

    _, err := d.Conn.WriteToUDP(p, addr)
    return err
}

func (d *DNSUtils) Inject(tun TUNPacket) ([]*dns.Msg, error){

    switch tun.GetCmd() {
    case TUN_CMD_DATA:

        t, ok := tun.(*TUNIPPacket)
        if !ok {
            return nil, fmt.Errorf("Invail Conversion\n")
        }
        return d.InjectIPPacket(uint16(t.Id), t.Payload)
    case TUN_CMD_CONNECT, TUN_CMD_KILL:

        // TODO
        msg, err := d.NewDNSPacket(tun)
        if err != nil {
            Error.Println(err)
            return nil, err
        }
        return []*dns.Msg{msg}, nil

    case TUN_CMD_RESPONSE:

        Error.Println("Inject for RESPONSE Not Implemented")
        // TODO
    default:
        return nil, fmt.Errorf("Invalid TUN CMD %s", tun.GetCmd())
    }

    return nil, fmt.Errorf("Not Implement\n")
}

/* Given a DNS Packet, Retrieve TUNPacket from it */
func (d *DNSUtils) Retrieve(dns *dns.Msg) (TUNPacket, error){

    /*
    switch tun.GetCmd() {
    case TUN_CMD_DATA:
        return d.InjectIPPacket(tun.Id, tun.Payload)
    case TUN_CMD_CREAT, TUN_CMD_KILL:
        // TODO
    case TUN_CMD_RESPONSE:
        // TODO
    default:
        return fmt.Errorf("Invalid TUN CMD %s", tun.GetCmd())
    }*/

    t := new(TUNCmdPacket)
    t.Cmd  = TUN_CMD_NONE
    t.User = 0

    return t, nil

    return t, fmt.Errorf("Not Implement\n")
}

/* Pack a DNS Packet to byte array */
/*
func (d *DNSUtils) Pack(*dns.Msg) ([]byte, error){

}*/

/* Given a byte array, Retrieve DNS Packet from it */
/*
func (d *DNSUtils) Unpack(b []byte) (*dns.Msg, error){

}*/

func (d *DNSUtils) injectToLabels(b []byte) ([]string, error) {

    encodedStr := base32.StdEncoding.EncodeToString(b)

	//decodedStr,_ := base32.StdEncoding.DecodeString(encodedStr)
	//	fmt.Printf("encodedStr %s\n", encodedStr)
	//	fmt.Printf("decodedStr %s\n", decodedStr)
	//	fmt.Printf("originStr  %s\n", string(raw))
	numLabels := len(encodedStr) / LABEL_SIZE
	labelsArr := make([]string, 0)

	for i := 0; i < numLabels; i++ {
		labelsArr = append(labelsArr, encodedStr[i*LABEL_SIZE:(i+1)*LABEL_SIZE])
	}
	// padding the last partially filled label
	if len(encodedStr)%LABEL_SIZE != 0 {
		lastLabel := encodedStr[numLabels*LABEL_SIZE:]
		lastLabel += strings.Repeat("_", (LABEL_SIZE - len(lastLabel)))
		labelsArr = append(labelsArr, lastLabel)
	}

	// padding 1-3 empty labels to labelsArr so that len(labelsArr) % 4 == 0
	for {
		if len(labelsArr)%4 == 0 {
			break
		}
		labelsArr = append(labelsArr, strings.Repeat("_", LABEL_SIZE))
	}

	//fmt.Printf("numLabels: %d, numDnsMsg: %d\n", numLabels, len(labelsArr)/4)
	return labelsArr, nil
}

func (d *DNSUtils) InjectIPPacket(id uint16, b []byte) ([]*dns.Msg, error){

    msgs := make([]*dns.Msg, 0)

    if d.Kind == DNS_Client {
        // Client: Insert into DNS Query

        labels, err := d.injectToLabels(b)
        if err != nil {
            return nil, err
        }

        cmdStr  := TUN_CMD_DATA
        idStr := strconv.FormatUint(uint64(id), 10)

        for i := 0; i < len(labels)/4; i++ {

            currLabels := labels[i*4 : (i+1)*4]
            encodedStr := strings.Join(currLabels, ".")
            var mf string = "1"
            if i == len(labels)/4-1 {
                mf = "0"
            }

            idxStr := strconv.Itoa(i)
            domainLabels := []string{encodedStr, idStr, mf, idxStr, string(cmdStr), d.TopDomain}

            domain := strings.Join(domainLabels, ".")

            if len(msgs) >= 253 {
                return nil, fmt.Errorf("Packed Msg Size %d > 253\n", msgs)
            }

            currMsg := new(dns.Msg)
            currMsg.SetQuestion(domain, dns.TypeA)
            currMsg.RecursionDesired = true

            //Debug.Println(currMsg.String())
            msgs = append(msgs, currMsg)
        }
    }else{
        // Server TODO: insert into TXT records
    }
	return msgs, nil
}

/* inject ip packet */
func (d *DNSUtils) InjectAndSendTo(b []byte, addr *net.UDPAddr) error {

    ippkt := new(ip.IPPacket)
    err := ippkt.Unmarshal(b)
    if err != nil {
        return err
    }

    id := ippkt.Header.Id

    t := new(TUNIPPacket)
    t.Cmd = TUN_CMD_DATA
    t.Id = int(id)
    t.More = false
    t.Offset = 0
    t.Payload = b

    msgs, err := d.Inject(t)
    if err != nil {
        return err
    }

    for i:=0 ; i<len(msgs) ; i++{

        pkt, err := msgs[i].Pack()
        if err != nil {
            return err
        }

        err = d.SendTo(addr, pkt)
        if err != nil {
            return err
        }
    }
    return nil
}
