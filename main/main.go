package main

import (
	"Kademlia---P2P-DFS/kdmlib"
	"fmt"
	"time"
)

func main() {
	id1 := "00001"
	id2 := "00010"
	id3 := "00100"
	id4 := "01000"
	routingT := kdmlib.CreateAllWorkersForRoutingTable(20, 5, 5, "00000")
	//routingT := kdmlib.CreateAllWorkersForRoutingTable(20,160,5,kdmlib.GenerateRandID())
	routingT.GiveOrder(kdmlib.OrderForRoutingTable{kdmlib.ADD, kdmlib.AddressTriple{"127.0.0.1", "9000", id1}, false})
	routingT.GiveOrder(kdmlib.OrderForRoutingTable{kdmlib.ADD, kdmlib.AddressTriple{"127.0.0.1", "9000", id2}, false})
	routingT.GiveOrder(kdmlib.OrderForRoutingTable{kdmlib.ADD, kdmlib.AddressTriple{"127.0.0.1", "9000", id3}, false})
	routingT.GiveOrder(kdmlib.OrderForRoutingTable{kdmlib.ADD, kdmlib.AddressTriple{"127.0.0.1", "9000", id4}, false})
	time.Sleep(time.Second)

	nw := kdmlib.InitializeNetwork(5, 12000, routingT)
	fmt.Println(nw)

}

func StartKademlia() {
	nodeId := kdmlib.GenerateRandID()
	rt := kdmlib.CreateAllWorkersForRoutingTable(kdmlib.K, kdmlib.IDLENGTH, 5, nodeId)
	nw := kdmlib.InitializeNetwork(3, 12000, rt)
	kdmlib.NewKademliaInstance(nw, nodeId, kdmlib.ALPHA, kdmlib.K, rt)
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
