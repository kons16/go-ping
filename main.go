package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

type Type uint8

const (
	ECHO_REPLY Type = 0
	ECHO       Type = 8
)

type IcmpHdr struct {
	Type     Type
	Code     uint8
	Checksum uint16
	ID       uint16
	Seq      uint16
	Data     []byte
}

func (m *IcmpHdr) Marshal() []byte {
	b := make([]byte, 8+len(m.Data))
	b[0] = byte(m.Type)
	b[1] = byte(m.Code)
	b[2] = 0
	b[3] = 0
	binary.BigEndian.PutUint16(b[4:6], m.ID)
	binary.BigEndian.PutUint16(b[6:8], m.Seq)
	copy(b[8:], m.Data)
	cs := checkSum(b)
	b[2] = byte(cs >> 8)
	b[3] = byte(cs)
	return b
}

// checkSum はチェックサムの計算を行う
func checkSum(b []byte) uint16 {
	count := len(b)
	sum := uint32(0)
	for i := 0; i < count-1; i += 2 {
		sum += uint32(b[i])<<8 | uint32(b[i+1])
	}
	if count&1 != 0 {
		sum += uint32(b[count-1]) << 8
	}
	for (sum >> 16) > 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}
	return ^(uint16(sum))

}

// sendPing は ping を送信する
func sendPing(conn net.Conn, name string, len int, sqc uint16, sendTime *time.Time) int {
	tb, err := time.Now().MarshalBinary()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Time.MarshalBinary:", err)
		os.Exit(1)
	}

	m := IcmpHdr{
		Type: ECHO,
		Code: 0,
		ID:   uint16(os.Getpid() & 0xffff),
		Seq:  sqc,
		Data: tb,
	}

	mb := m.Marshal()
	_, err = conn.Write(mb)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Write:", err)
		os.Exit(1)
	}

	return 0
}

// checkPacket は受信したパケットを確認する
func checkPacket(b []byte) int {
	hlen := int(b[0]&0x0f) << 2

	b = b[hlen:]
	m := &IcmpHdr{
		Type:     Type(b[0]),
		Code:     uint8(b[1]),
		Checksum: uint16(binary.BigEndian.Uint16(b[2:4])),
		ID:       uint16(binary.BigEndian.Uint16(b[4:6])),
		Seq:      uint16(binary.BigEndian.Uint16(b[6:8])),
	}
	m.Data = make([]byte, len(b)-8)
	copy(m.Data, b[8:])

	// fmt.Fprintf("%d bytes from %s: icmp_seq=%d ttl=%d time=%.2f ms\n",nbytes-iph->ihl*4,inet_ntoa(from->sin_addr),sqc,*ttl,*diff*1000.0)
	fmt.Println("seq = ", m.Seq)
	return 0
}

// recvPing は ping 応答を受信する
func recvPing(conn net.Conn, len int, sqc uint16, sendTime *time.Time, timeOutSec int) int {
	rb := make([]byte, 100)
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))

	n, err := conn.Read(rb)
	if err != nil {
		return -2000
	}

	ret := checkPacket(rb[:n])
	if ret == 0 {
		return 0
	} else {
		return 1
	}
}

// pingCheck は ping の送受信を行う
// name 疎通確認先, len ICMPパケット長, times ICMP送受信繰り返し回数, timeOutSec 応答受信待ち秒数
func pingCheck(ipStr string, len int, times int, timeOutSec int) (int, error) {
	var sendTime time.Time
	var ret int
	var total, totalNo int

	// ソケットを作成
	conn, err := net.Dial("ip4:1", ipStr)
	if err != nil {
		fmt.Println("aaa")
		fmt.Sprintf("%s", err)
		return -1, err
	}
	defer conn.Close()

	for i := 0; i < times; i++ {
		// ICMPエコーリクエストを送信する
		ret = sendPing(conn, ipStr, len, uint16(i+1&0xffff), &sendTime)
		if ret == 0 {
			// ICMPエコーリプライを受信する
			ret = recvPing(conn, len, uint16(i+1&0xffff), &sendTime, timeOutSec)
			if ret >= 0 {
				total += ret
				totalNo++
			}
		}

		time.Sleep(1)
	}

	if totalNo > 0 {
		return total / totalNo, nil
	} else {
		return -1, nil
	}
}

func main() {
	flag.Parse()
	ipStr := flag.Args()[0]

	ret, err := pingCheck(ipStr, 64, 5, 1)
	if err != nil {
		fmt.Sprintf("%s", err)
		os.Exit(1)
	}

	if ret >= 0 {
		fmt.Printf("RTT:%d\n", ret)
		os.Exit(0)
	} else {
		fmt.Printf("error:%d\n", ret)
		os.Exit(1)
	}
}
