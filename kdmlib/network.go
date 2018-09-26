package kdmlib

import (
	"Kademlia---P2P-DFS/github.com/golang/protobuf/proto"
	pb "Kademlia---P2P-DFS/kdmlib/proto_config"
	"fmt"
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
	ServerAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(port))
	CheckError(err)

	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	buf := make([]byte, 1024)

	quit := make(chan struct{})
	go Listen(buf, ServerConn, ServerAddr, quit)
	<-quit
}

func Listen(buf []byte, ServerConn *net.UDPConn, ServerAddr *net.UDPAddr, quit chan struct{}) { //TODO: ADD worker pools
	defer ServerConn.Close()
	for {
		n, _, err := ServerConn.ReadFromUDP(buf) //add addr

		packet := &pb.Package{}
		err = proto.Unmarshal(buf[0:n], packet)
		if err != nil {
			fmt.Println("Error: ", err)
		}

		Handler(packet, ServerConn, ServerAddr)
	}
	quit <- struct{}{}
}
func SendSomething(contact AddressTriple, Conn *net.UDPConn) {

	ServerAddr, err := net.ResolveUDPAddr("udp", contact.Ip+":"+contact.Port)
	CheckError(err)
	pack := &pb.Package{Id: "Request", Type: "SendSomething", Message: "TEST", Time: time.Now().String()}
	//pack := CreatePackage("Request", "SendSomething", "TEST", time.Now().String()) //<- test buffers directly
	fmt.Println("Sending a:", pack.Id, "with type:", pack.Type, "to:", ServerAddr)
	SendData(PackageToMarshal(pack), Conn, ServerAddr)
	//defer Conn.Close()
}
func (network *Network) SendPingMessage(contact AddressTriple) {

}
func (network *Network) SendFindContactMessage(target string, contact *AddressTriple, returnChannel chan interface{}) {
	// TODO
}
func (network *Network) SendFindDataMessage(hash string, contact *AddressTriple, returnChannel chan interface{}) {
	// TODO
}
func (network *Network) SendStoreMessage(data []byte) {
	// TODO
}
func Handler(packet *pb.Package, serverConn *net.UDPConn, addr *net.UDPAddr) {

	switch packet.Id {
	case "Request":
		switch packet.Type {
		case "SendSomething":
			fmt.Println("Received a:", packet.Id, "with type:", packet.Type, "from: ", addr)
			//Create appropriate pack with belonging data and ship it out
			pack := &pb.Package{Id: "Return", Type: "ReturnSomething", Message: "TEST", Time: time.Now().String()}
			fmt.Println("Returning a:", pack.Id, "with type:", pack.Type, "to: ", addr)
			SendData(PackageToMarshal(pack), serverConn, addr)
			break
		case "Ping":
			break
		case "FindContact":
			break
		case "FindData":
			break
		case "Store":
			break
		}
		break
	case "Return":
		switch packet.Type {
		case "ReturnSomething":
			fmt.Println("Received a: ", packet.Id, "with type:", packet.Type, "from: ", addr)
			break
		case "Ping":
			break
		case "Contact":
			break
		case "Data":
			break
		case "Store":
			break
		}
		break
	}
}

func SendData(data []byte, Conn *net.UDPConn, addr *net.UDPAddr) {

	buf := []byte(data)
	_, err := Conn.WriteToUDP(buf, addr)
	if err != nil {
		fmt.Println(data, err)
	}
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

//Kademlia Documentations
/*
	This RPC involves one node sending a PING message to another, which presumably replies with a PONG.
	This has a two-fold effect: the recipient of the PING must update the bucket corresponding to the sender;
	and, if there is a reply, the sender must update the bucket appropriate to the recipient.
	All RPC packets are required to carry an RPC identifier assigned by the sender and echoed in the reply.
	This is a quasi-random number of length B (160 bits).
*/ //Ping
/*
The FIND_NODE RPC includes a 160-bit key. The recipient of the RPC returns up to k triples (IP address, port, nodeID)
for the contacts that it knows to be closest to the key.
The recipient must return k triples if at all possible.
It may only return fewer than k if it is returning all of the contacts that it has knowledge of.
This is a primitive operation, not an iterative one.
*/ //Find_Contact
/*
A FIND_VALUE RPC includes a B=160-bit key. If a corresponding value is present on the recipient,
the associated data is returned. Otherwise the RPC is equivalent to a FIND_NODE and a set of k triples is returned.
This is a primitive operation, not an iterative one.
*/ //Find_data
/*
The sender of the STORE RPC provides a key and a block of data and requires that
the recipient store the data and make it available for later retrieval by that key.
This is a primitive operation, not an iterative one.
*/ //Store
