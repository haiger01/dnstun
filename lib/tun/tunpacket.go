package tun

import (
    "net"
)

const (
    TUN_CMD_CONNECT  byte = 'c'
    TUN_CMD_RESPONSE byte = 'r'
    TUN_CMD_DATA     byte = 'd'
    TUN_CMD_KILL     byte = 'k'
)

type TUNPacket interface {
    GetCmd() byte
}

type TUNCmdPacket struct {
    Cmd byte
}

type TUNResponsePacket struct {
    Cmd     byte
    Server   *net.IPAddr
    Client   *net.IPAddr
}

type TUNIPPacket struct {
    Cmd     byte
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

