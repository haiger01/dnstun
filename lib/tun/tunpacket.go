package tun

import (
    "net"
)

const (
    TUN_CMD_CONNECT  byte = 'c'
    TUN_CMD_RESPONSE byte = 'r'
    TUN_CMD_DATA     byte = 'd'
    TUN_CMD_KILL     byte = 'k'
    TUN_CMD_EMPTY    byte = 'e' // empty packet, with user id,
                                // just for server to have more dns id
    TUN_CMD_NONE     byte = 'n' // no user id, a normal dns request
)

type TUNPacket interface {
    GetCmd()    byte

    /* The Physical UDP Address for an incoming packet 
       may change over time, e.g. using different middle 
       DNS Server. By using User field to identify the source
       Of a TUN Packet */
    GetUser()   int
}

type TUNCmdPacket struct {
    Cmd byte
    User int
}

type TUNResponsePacket struct {
    Cmd     byte
    User    int
    Server   *net.IPAddr
    Client   *net.IPAddr
}

type TUNIPPacket struct {
    Cmd     byte
    User    int
    Id      int
    Offset  int
    More    bool
    Payload []byte
}
func (t *TUNCmdPacket) GetCmd() byte{
    return t.Cmd
}
func (t *TUNResponsePacket) GetCmd() byte{
    return TUN_CMD_RESPONSE
}
func (t *TUNIPPacket) GetCmd() byte{
    return TUN_CMD_DATA
}
func (t *TUNCmdPacket) GetUser() int{
    return t.User
}
func (t *TUNResponsePacket) GetUser() int{
    return t.User
}
func (t *TUNIPPacket) GetUser() int{
    return t.User
}

/*
func (t *TUNPacket) Unpack(domain string) (*TUNPacket, error){

    // TODO: TUNCmdPacket, TUNResponsePacket

    labels := strings.Split(name, ".")

	// labels[0]...labels[3]: dlabel(54)
	// labels[4]: identification(5)
	// labels[5]: MF(1)
	// labels[6]: idx(4)
	// labels[7]: cmd(2)
	// labels[8:]: b.jannotti.com(14)
	// total: 54*4+3 + 5+1+1+1+4+1+2+1+14 = 249
	_id, _ := strconv.Atoi(labels[4])
	_mf, _ := strconv.Atoi(labels[5])
	_idx, _ := strconv.Atoi(labels[6])
	_cmd, _ := strconv.Atoi(labels[7])
	outPacket := &TUNIPPacket{
		Cmd: _cmd,
		Id:  _id,
		Idx: _idx,
	}
	if _mf == 1 {
		outPacket.More = true
	}
	raw := labels[0] + labels[1] + labels[2] + labels[3]

	outPacket.EncodedStr = raw
	return outPacket, nil
}
*/

