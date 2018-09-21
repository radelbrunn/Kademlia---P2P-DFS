package src

import (
	"fmt"
	"net"
	"os"
	"time"
)

type Network struct {
}

func UDPConnection() {
	/* Lets prepare a address at any address at port 10002*/
	ServerAddr, err := net.ResolveUDPAddr("udp", ":10002")
	CheckError(err)

	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)

	quit := make(chan struct{})
	go SendMessage()
	go Listen(ServerConn, quit)
	<-quit
}

func Listen(ServerConn *net.UDPConn, quit chan struct{}) {
	buf := make([]byte, 1024)
	defer ServerConn.Close()
	for {

		n, addr, err := ServerConn.ReadFromUDP(buf)
		fmt.Println("Received ", string(buf[0:n]), " from ", addr)

		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
	quit <- struct{}{}
}

func SendMessage() {
	ServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:10002")
	CheckError(err)

	LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	CheckError(err)

	Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	CheckError(err)
	SendPingMessage(Conn)
	defer Conn.Close()
}
func SendPingMessage(Conn *net.UDPConn) {

	for {
		msg := "PING"
		buf := []byte(msg)
		_, err := Conn.Write(buf)

		if err != nil {
			fmt.Println(msg, err)
		}
		time.Sleep(time.Second * 1)
	}
}

func (network *Network) SendFindContactMessage(contact *Contact) {
	// TODO
}

func (network *Network) SendFindDataMessage(hash string) {
	// TODO
}

func (network *Network) SendStoreMessage(data []byte) {
	// TODO
}
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}
