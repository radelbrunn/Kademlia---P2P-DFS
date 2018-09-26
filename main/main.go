package main

import (
	"Kademlia---P2P-DFS/github.com/golang/protobuf/proto"
	"Kademlia---P2P-DFS/kdmlib"
	pb "Kademlia---P2P-DFS/kdmlib/proto_config"
	"fmt"
	"net"
	"time"
)

func main() {
	StartKademlia()
	//networktest()
}

func StartKademlia() {
	id := kdmlib.GenerateRandID()
	port := 44444 //Need high port number
	nw := kdmlib.Initialize_Network(port)
	kademlia := kdmlib.NewKademliaInstance(nw, id, kdmlib.ALPHA, kdmlib.K)

	fmt.Println(kademlia)
}

func networktest() {

	id := kdmlib.GenerateIDFromHex("ffffffff00000000000000000000000000000000")
	//channel := make(chan interface{})
	contact := kdmlib.AddressTriple{"localhost", "9000", id}
	serverAddr, err := net.ResolveUDPAddr("udp", contact.Ip+":"+contact.Port)
	kdmlib.CheckError(err)
	serverConn, err := net.ListenUDP("udp", serverAddr)
	kdmlib.CheckError(err)
	defer serverConn.Close()
	buf := make([]byte, 4096)

	go kdmlib.SendSomething(contact, serverConn)

	for {
		n, _, err := serverConn.ReadFromUDP(buf) //add addr

		packet := &pb.Package{}
		err = proto.Unmarshal(buf[0:n], packet)
		if err != nil {
			fmt.Println("Error: ", err)
		}

		Handler(packet, serverConn, serverAddr)
	}

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
			kdmlib.SendData(kdmlib.PackageToMarshal(pack), serverConn, addr)
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

/* ROUTING TABLE TEST

func main() {
	channelIn := make(chan OrderForPinger, 100)
	channelOut := make(chan OrderForRoutingTable)

	mylist := list.New()
	mylist.PushFront(AddressTriple{"127.0.0.1", "1053", ""})

	routingtable := CreateRoutingTable(20, 4)
	go UpdateRoutingTableWorker(routingtable, channelOut, "0000", 20, channelIn)
	addressToAdd := AddressTriple{"127.0.0.1", "8080", "0010"}
	addressToAdd2 := AddressTriple{"127.0.0.1", "8080", "0011"}
	channelOut <- OrderForRoutingTable{ADD, addressToAdd, false}
	channelOut <- OrderForRoutingTable{ADD, addressToAdd2, false}

	time.Sleep(time.Second)


	//fmt.Println(findKClosest(*routingtable.routingTable, "0100", 2))
	fmt.Println((*routingtable.routingTable)[2].Front().Value)
	fmt.Println((*routingtable.routingTable)[2].Front().Next().Value)
}
*/
