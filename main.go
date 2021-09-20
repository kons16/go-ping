package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

// checkSum はチェックサムの計算を行う
func checkSum() {}

// sendPing は ping を送信する
func sendPing(conn net.Conn, name string, len int, sqc int, sendTime *time.Time) int {}

// checkPacket は受信したパケットを確認する
func checkPacket() {}

// recvPing は ping 応答を受信する
func recvPing(conn net.Conn, len int, sqc int, sendTime *time.Time, timeOutSec int) int {}

// pingCheck は ping の送受信を行う
// name 疎通確認先, len ICMPパケット長, times ICMP送受信繰り返し回数, timeOutSec 応答受信待ち秒数
func pingCheck(name string, len int, times int, timeOutSec int) (int, error) {
	var ip string
	var sendTime time.Time
	var ret int
	var total, totalNo int

	// ソケットを作成
	conn, err := net.Dial("ip4:1", ip)
	if err != nil {
		fmt.Sprintf("%s", err)
		return -1, err
	}
	defer conn.Close()

	for i := 0; i < times; i++ {
		// ICMPエコーリクエストを送信する
		ret = sendPing(conn, name, len, i+1, &sendTime)
		if ret == 0 {
			// ICMPエコーリプライを受信する
			ret = recvPing(conn, len, i+1, &sendTime, timeOutSec)
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
	name := flag.Args()[0]

	ret, err := pingCheck(name, 64, 5, 1)
	if err != nil {
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
