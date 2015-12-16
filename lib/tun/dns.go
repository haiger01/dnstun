package tun

import (
	"../ip"
	"../tonnerre/golang-dns"
	"encoding/base32"
	"fmt"
	"net"
	"strconv"
	"strings"
)

const (
	DNS_Client int = 0
	DNS_Server int = 1
)

type DNSUtils struct {
	Kind      int
	Conn      *net.UDPConn
	TopDomain string
	LAddr     *net.UDPAddr
	LDns      *net.UDPAddr
}

func NewDNSClient(laddrstr, ldnsstr, topDomain string) (*DNSUtils, error) {

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

func NewDNSServer(laddrstr, topDomain string) (*DNSUtils, error) {

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

func (d *DNSUtils) NewDNSPacket(t TUNPacket) (*dns.Msg, error) {

	switch t.GetCmd() {
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

func (d *DNSUtils) Send(p []byte) error {
	if d.Kind != DNS_Client {
		return fmt.Errorf("Send: Only used by Client\n")
	}
	_, err := d.Conn.WriteToUDP(p, d.LDns)
	return err
}

func (d *DNSUtils) SendTo(addr *net.UDPAddr, p []byte) error {

	_, err := d.Conn.WriteToUDP(p, addr)
	return err
}

func (d *DNSUtils) Reply(msg *dns.Msg, tun TUNPacket, paddr *net.UDPAddr) error {

	var msgs []*dns.Msg
	var err error
	switch tun.GetCmd() {
	case TUN_CMD_RESPONSE:
		msgs, err = d.Inject(tun)
		if err != nil {
			Error.Println(err)
			return err
		}
	case TUN_CMD_ACK:
		msgs, err = d.Inject(tun)
		if err != nil {
			return err
		}
	case TUN_CMD_DATA:
		// Encode

		Error.Println("DNS Reply: Not Implemented")

	default:
		return fmt.Errorf("DNS Reply: Invalid TUN Cmd")
	}
	fmt.Printf("dns response to send: \n")
	for _, msg := range msgs {

		fmt.Println(msg.String())
		binary, err := msg.Pack()
		if err != nil {
			return err
		}
		err = d.SendTo(paddr, binary)
		if err != nil {
			return err
		}
	}
	return nil
}

// This function works better for Client's purpose, as Server need Client's DNS request Msg to call SetReply
func (d *DNSUtils) Inject(tun TUNPacket) ([]*dns.Msg, error) {

	msgs := make([]*dns.Msg, 0)

	switch tun.GetCmd() {
	case TUN_CMD_DATA:

		t, ok := tun.(*TUNIPPacket)
		if !ok {
			return nil, fmt.Errorf("Invail Conversion\n")
		}
		return d.InjectIPPacket(uint16(t.Id), t.Payload)
	case TUN_CMD_CONNECT:
		msg, err := d.NewDNSPacket(tun)
		if err != nil {
			Error.Println(err)
			return nil, err
		}
		msgs = append(msgs, msg)
		return msgs, nil

	case TUN_CMD_EMPTY:
		msg := new(dns.Msg)
		labels := []string{string(tun.(*TUNCmdPacket).UserId), string(TUN_CMD_EMPTY), d.TopDomain}
		domain := strings.Join(labels, ".")
		msg.SetQuestion(domain, dns.TypeTXT)
		msg.RecursionDesired = true
		msgs = append(msgs, msg)
		return msgs, nil

	case TUN_CMD_KILL:
		Error.Println("Inject for CMD_KILL not implemented")
		return nil, nil

	case TUN_CMD_RESPONSE, TUN_CMD_ACK:
		var replyStr string
		var ans dns.RR
		var err error
		reply := new(dns.Msg)
		reply.Answer = make([]dns.RR, 1)
		ans.(*dns.TXT).Txt = make([]string, 3)
		if tun.GetCmd() == TUN_CMD_RESPONSE {
			tunPkt, ok := tun.(*TUNResponsePacket)
			if !ok {
				return nil, fmt.Errorf("error casting to TUNResponsePacket\n")
			}
			domain := tunPkt.Request.Question[0].Name
			ans, err = dns.NewRR(domain + " 0 IN TXT xx")
			if err != nil {
				return nil, err
			}
			reply.SetReply(tunPkt.Request)
			serverIpStr := strings.Replace(tunPkt.Server.String(), ".", "_", -1)
			clientIpStr := strings.Replace(tunPkt.Client.String(), ".", "_", -1)
			replyDomains := []string{string(TUN_CMD_RESPONSE), strconv.Itoa(tunPkt.UserId), serverIpStr, clientIpStr}
			replyStr = strings.Join(replyDomains, ".")
		} else {
			tunPkt, ok := tun.(*TUNAckPacket)
			if !ok {
				return nil, fmt.Errorf("error casting to TUNAckPacket\n")
			}
			domain := tunPkt.Request.Question[0].Name
			ans, err = dns.NewRR(domain + " 0 IN TXT xx")
			if err != nil {
				return nil, err
			}
			reply.SetReply(tunPkt.Request)
			replyStr = string(TUN_CMD_ACK)
		}
		ans.(*dns.TXT).Txt[0] = replyStr
		reply.Answer[0] = ans
		msgs = append(msgs, reply)
		return msgs, nil
	default:
		return nil, fmt.Errorf("Invalid TUN CMD %s", tun.GetCmd())
	}

	return nil, fmt.Errorf("Not Implement\n")
}

/* Given a DNS Packet, Retrieve TUNPacket from it */
func (d *DNSUtils) Retrieve(in *dns.Msg) (TUNPacket, error) {

	if len(in.Answer) > 0 {
		// dns packet sent from DNSServer
		ans, ok := in.Answer[0].(*dns.TXT)
		if !ok {
			return nil, fmt.Errorf("unexpected reply RR record, not TXT\n")
		}
		cmdDomains := strings.Split(ans.Txt[0], ".")
		cmd := byte(cmdDomains[0][0])
		var err error
		switch cmd {
		case TUN_CMD_RESPONSE:
			t := new(TUNResponsePacket)
			t.Cmd = TUN_CMD_RESPONSE
			t.UserId, err = strconv.Atoi(cmdDomains[1])
			if err != nil {
				return nil, err
			}
			serverIpStr := strings.Replace(cmdDomains[2], "_", ".", -1)
			clientIpStr := strings.Replace(cmdDomains[3], "_", ".", -1)
			t.Server, err = net.ResolveIPAddr("ip", serverIpStr)
			if err != nil {
				return nil, err
			}
			t.Client, err = net.ResolveIPAddr("ip", clientIpStr)
			if err != nil {
				return nil, err
			}
			return t, nil

		default:
			return nil, fmt.Errorf("TUN_CMD %s from DNSServer not implemented \n", string(cmd))
		}

		return nil, fmt.Errorf("DNSUtils.Retrieve should not be here")

	} else {
		// dns packet sent from DNSClient
		t := new(TUNCmdPacket)
		r := in.Question[0]
		domains := strings.Split(r.Name[:len(r.Name)-1], ".") // trim the last '.'from "b.jannotti.com."[-1]
		n := len(domains)
		if n < 4 {
			return nil, fmt.Errorf("unexpecetd domain name format %s\n", r.Name)
		}
		cmd := byte(domains[n-4][0])
		if cmd != TUN_CMD_CONNECT && n < 5 {
			return nil, fmt.Errorf("unexpecetd domain name format %s\n", r.Name)
		}
		t.Cmd = cmd
		switch cmd {
		case TUN_CMD_CONNECT:
			t.UserId = -1 // has not been allocated by DNSServer
		default:
			var err error
			t.UserId, err = strconv.Atoi(domains[n-5])
			if err != nil {
				return nil, fmt.Errorf("cannot parse UserId from %s\n", domains[n-5])
			}
		}
		return t, nil
	}

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

func (d *DNSUtils) InjectIPPacket(id uint16, b []byte) ([]*dns.Msg, error) {

	msgs := make([]*dns.Msg, 0)

	if d.Kind == DNS_Client {
		// Client: Insert into DNS Query

		labels, err := d.injectToLabels(b)
		if err != nil {
			return nil, err
		}

		cmdStr := TUN_CMD_DATA
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
	} else {
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

	for i := 0; i < len(msgs); i++ {

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
