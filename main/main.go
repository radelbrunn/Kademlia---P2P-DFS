package main

import (
	pb "Kademlia---P2P-DFS/kdmlib/proto_config"
	"Kademlia/src_code"
	"fmt"
	"github.com/golang/protobuf/proto"
	"net"
)

func main() {
	//Test. Send message as a package with marshal.
	// This will eventually be modified and moved to network_test.go once its created
	id := src_code.NewKademliaID("ffffffff00000000000000000000000000000000")
	//channel := make(chan interface{}) //todo: learn how to use channels proper
	contact := src_code.NewContact(id, "localhost:9000")
	serverAddr, err := net.ResolveUDPAddr("udp", contact.Address)
	src_code.CheckError(err)
	serverConn, err := net.ListenUDP("udp", serverAddr)
	src_code.CheckError(err)
	defer serverConn.Close()
	buf := make([]byte, 4096)

	go src_code.SendMessageAsAPackageTest(contact)
	for {
		n, addr, err := serverConn.ReadFromUDP(buf)

		packet := &pb.Package{}
		err = proto.Unmarshal(buf[0:n], packet)
		fmt.Println("Received data from: ", addr, "DATA :=> ", packet.Id, " : ", packet.Message, ": ", packet.Time)

		if err != nil {
			fmt.Println("Error: ", err)
		}
		return
	}
}

//func main() {
//	contact := kdmlib.NewContact(kdmlib.NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000")
//	fmt.Print(contact.String())
//}

//func main() {
//	channelIn := make(chan OrderForPinger, 100)
//	channelOut := make(chan OrderForRoutingTable)
//
//	mylist := list.New()
//	mylist.PushFront(AddressTriple{"127.0.0.1", "1053", ""})
//
//	routingtable := CreateRoutingTable(20, 4)
//	go UpdateRoutingTableWorker(routingtable, channelOut, "0000", 20, channelIn)
//	addressToAdd := AddressTriple{"127.0.0.1", "8080", "0010"}
//	addressToAdd2 := AddressTriple{"127.0.0.1", "8080", "0011"}
//	channelOut <- OrderForRoutingTable{ADD, addressToAdd, false}
//	channelOut <- OrderForRoutingTable{ADD, addressToAdd2, false}
//
//	time.Sleep(time.Second)
//
//
//	//fmt.Println(findKClosest(*routingtable.routingTable, "0100", 2))
//	fmt.Println((*routingtable.routingTable)[2].Front().Value)
//	fmt.Println((*routingtable.routingTable)[2].Front().Next().Value)
//}
