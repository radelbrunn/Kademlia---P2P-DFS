package kdmlib

import (
	pb "Kademlia---P2P-DFS/kdmlib/proto_config"
	"fmt"
	"github.com/golang/protobuf/proto"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

//Work in progress. Much will change!
type Network struct {
	//TODO: add stuff. mux, nodes? eventually file network?
}

func Initialize_Network(port int) *Network {
	//TODO: Add network stuff
	network := &Network{}
	UDPConnection(port)
	return network
}
func UDPConnection(port int) { //TODO: learn how to properly use channels
	/* Lets prepare a address at any address at port "port"*/
	ServerAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(port))
	CheckError(err)

	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	buf := make([]byte, 1024)

	quit := make(chan struct{})
	go Listen(buf, ServerConn, quit)
	<-quit
}

func Listen(buf []byte, ServerConn *net.UDPConn, quit chan struct{}) {
	defer ServerConn.Close()
	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)
		//TODO: add unmarshaling - already done in test file but need modifications
		fmt.Println("Received ", string(buf[0:n]), " from ", addr)
		//TODO: add message request handlers
		if err != nil {
			fmt.Println("Error: ", err)
		}
	}
	quit <- struct{}{}
}
func SendMessageAsAPackageTest(contact Contact) {
	messageID := NewRandomKademliaID()

	ServerAddr, err := net.ResolveUDPAddr("udp", contact.Address)
	CheckError(err)

	LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	CheckError(err)

	Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	CheckError(err)

	pack := CreatePackege(messageID.String(), "My_message", time.ANSIC)
	SendData(PackageToMarshal(pack), Conn)
	defer Conn.Close()
}
func (network *Network) SendPingMessage(contact Contact) {

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
func SendData(data []byte, Conn *net.UDPConn) {

	for {
		buf := []byte(data)
		_, err := Conn.Write(buf)

		if err != nil {
			fmt.Println(data, err)
		}
		time.Sleep(time.Second * 1)
	}
}
func CreatePackege(id string, msg string, time string) *pb.Package {
	pack := pb.Package{
		Id:      id,
		Message: msg,
		Time:    time,
	}
	return &pack
}
func PackageToMarshal(pack *pb.Package) []byte {
	data, err := proto.Marshal(pack)
	if err != nil {
		log.Fatal("marshalling error: ", err)
	}
	return data
}
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}
