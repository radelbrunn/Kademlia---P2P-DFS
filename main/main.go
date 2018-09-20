package main

import (
	"Kademlia---P2P-DFS/kdmlib"
	"fmt"
)

func main() {
	StartKademlia()
}

func StartKademlia() {
	id := kdmlib.GenerateRandID()
	ip := "127.0.0.1"
	port := 87
	nw := kdmlib.NewNetwork(ip, port)
	kademlia := kdmlib.NewKademliaInstance(nw, id, kdmlib.ALPHA, kdmlib.K)

	fmt.Println(kademlia)
}

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
