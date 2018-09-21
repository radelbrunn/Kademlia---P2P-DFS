package main

import (
	"fmt"
	"net"
	"time"
)

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func main() {
	ServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:10002")
	CheckError(err)

	LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	CheckError(err)

	Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	CheckError(err)
	SendPingMessage(Conn)
	SendFindContactMessage(Conn)
	defer Conn.Close()
}
func SendPingMessage(Conn *net.UDPConn) {

	//for {
	msg := "PING"
	buf := []byte(msg)
	_, err := Conn.Write(buf)

	if err != nil {
		fmt.Println(msg, err)
	}
	time.Sleep(time.Second * 1)
	//}
}
func SendFindContactMessage(Conn *net.UDPConn) {

	//for {
	msg := "FIND"
	buf := []byte(msg)
	_, err := Conn.Write(buf)

	if err != nil {
		fmt.Println(msg, err)
	}
	time.Sleep(time.Second * 1)
	//}
}
