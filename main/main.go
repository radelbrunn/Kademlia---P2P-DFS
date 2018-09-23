package main

import (
	"Kademlia---P2P-DFS/github.com/golang/protobuf/proto"
	"Kademlia---P2P-DFS/kdmlib"
	pb "Kademlia---P2P-DFS/kdmlib/proto_config"
	"fmt"
	"net"
)

func main() {
	networktest()
	//ID := kdmlib.GenerateRandID()
	//kademlia := kdmlib.NewKademliaInstance(kdmlib.InitializeNetwork(9000), ID, kdmlib.ALPHA, kdmlib.K)
	//fmt.Println(kademlia)
}

func networktest() {

	id := kdmlib.GenerateIDFromHex("ffffffff00000000000000000000000000000000")
	//channel := make(chan interface{})
	contact := kdmlib.AddressTriple{"localhost", "9000", id}
	serverAddr, err := net.ResolveUDPAddr("udp", contact.Ip+":"+contact.Port)
	kdmlib.CheckError(err)
	serverConn, err := net.ListenUDP("udp", serverAddr)
	kdmlib.CheckError(err)
	//defer serverConn.Close()
	buf := make([]byte, 4096)

	go kdmlib.SendSomething(contact, serverConn)

	for {
		n, _, err := serverConn.ReadFromUDP(buf) //add addr

		packet := &pb.Container{}
		err = proto.Unmarshal(buf[0:n], packet)
		if err != nil {
			fmt.Println("Error: ", err)
		}

		Handler(packet, serverConn, serverAddr)
	}

}
func Handler(container *pb.Container, serverConn *net.UDPConn, addr *net.UDPAddr) {
	switch container.REQUEST_TYPE {
	case kdmlib.Request:
		switch container.REQUEST_ID {
		case kdmlib.Ping:
			//Create appropriate container with belonging data and ship it out
			//ID := kdmlib.node.nodeID, ID := network.node.nodeID
			fmt.Println("boop")
			Info := &pb.RETURN_PING{ID: "ME!"}
			Container := &pb.Container_ReturnPing{ReturnPing: Info}
			Data := &pb.Container{REQUEST_TYPE: kdmlib.Return, REQUEST_ID: kdmlib.Ping, Attachment: Container}
			kdmlib.Sendstuff(kdmlib.EncodeContainer(Data), serverConn, addr)
			fmt.Println("beep")
			break
		case kdmlib.FindContact:
			break
		case kdmlib.FindData:
			break
		case kdmlib.Store:
			break
		}
		break
	case kdmlib.Return:
		switch container.REQUEST_ID {
		case kdmlib.Ping:
			fmt.Println("boop")
			fmt.Println("repeat!")
			break
		case kdmlib.FindContact:
			break
		case kdmlib.FindData:
			break
		case kdmlib.Store:
			break
		}
		break
	}
}
