package tun

type Tunnel struct {
    name    string
    conn    *water.Interface
}

func NewTunnel(name string) (*Tunnel, error){

    t := new(Tunnel)
    t.name = name
    t.conn, err = water.NewTUN(name)
    if err != nil {
        return err
    }
}

func (t *Tunnel) Write(p []byte) error{

    n, err := t.ifce.Write(pkt)
    if err != nil {
        return err
    }
    if n != len(pkt){
        return fmt.Errorf("Short write %d, should be %d", n, len(pkt))
    }
}

func (t *Tunnel) Save(buffer map[int][]byte, tun *TUNIPPacket) error{

    if tun.Offset == 0 && tun.More == false {
        pkt := tun.Payload
        t.Write(pkt)   // send to upper layer
        return nil
    }
    pkt, ok := buffer[tun.Id]
    if ok {
        if tun.Offset == len(pkt) {
            pkt := append(pkt, tun.Payload)
            if tun.More == false{
                t.Write(pkt)
                delete(buffer, tun.Id)
            }else{
                buffer[tun.Id] = pkt
            }
        }
    }else{
        buffer[tun.Id] = tun.Payload
    }
    return nil
}

func (t *Tunnel) Read(p []byte) (int, err){
    n, err := t.conn.Read(p)
    return n, err
}
