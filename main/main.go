package main

import (
	"Kademlia---P2P-DFS/kdmlib"
	"math/rand"
)

func main() {
	/*
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

		nw := kdmlib.InitializeNetwork(5, 12000, routingT, false)
		fmt.Println(nw)
	*/

	StartKademlia()

}

func StartKademlia() {
	nodeId := kdmlib.GenerateRandID(int64(rand.Intn(100)))
	//connect to Google Cloud Node
	rt := kdmlib.CreateAllWorkersForRoutingTable(kdmlib.K, kdmlib.IDLENGTH, 5, nodeId)
	nw := kdmlib.InitializeNetwork(3, 12000, rt, false)
	kdmlib.NewKademliaInstance(nw, nodeId, kdmlib.ALPHA, kdmlib.K, rt)

}
