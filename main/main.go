package main

import (
	"Kademlia---P2P-DFS/kdmlib"
	"math/rand"
)

func main() {

	ownId := "0000"

	id1 := "0001"
	id2 := "0010"
	id3 := "0100"
	id4 := "1000"
	id5 := "0011"
	routingT := kdmlib.CreateAllWorkersForRoutingTable(20, 4, 5, ownId)
	routingT.GiveOrder(kdmlib.OrderForRoutingTable{kdmlib.ADD, kdmlib.AddressTriple{"127.0.0.1", "9000", id1}, false})
	routingT.GiveOrder(kdmlib.OrderForRoutingTable{kdmlib.ADD, kdmlib.AddressTriple{"127.0.0.1", "9000", id2}, false})
	routingT.GiveOrder(kdmlib.OrderForRoutingTable{kdmlib.ADD, kdmlib.AddressTriple{"127.0.0.1", "9000", id3}, false})
	routingT.GiveOrder(kdmlib.OrderForRoutingTable{kdmlib.ADD, kdmlib.AddressTriple{"127.0.0.1", "9000", id4}, false})
	routingT.GiveOrder(kdmlib.OrderForRoutingTable{kdmlib.ADD, kdmlib.AddressTriple{"127.0.0.1", "9000", id5}, false})

	//time.Sleep(time.Second)

	//addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12000") //<-- try this address when testing!
	//answerChannel := make(chan interface{})

	nw := kdmlib.InitializeNetwork(5, 12000, routingT, ownId, true)
	kd := kdmlib.NewKademliaInstance(nw, ownId, kdmlib.ALPHA, kdmlib.K, routingT)
	kd.LookupContact(kdmlib.AddressTriple{"127.0.0.1", "9000", "1001"}.Id, false)
	//	kd.LookupContact("1001", false)

	//nw2 := kdmlib.InitializeNetwork(5, 22000, routingT,nodeId, false)
	//nw2.SendPing(addr, answerChannel)

}

func StartKademlia() {
	nodeId := kdmlib.GenerateRandID(int64(rand.Intn(100)))
	rt := kdmlib.CreateAllWorkersForRoutingTable(kdmlib.K, kdmlib.IDLENGTH, 5, nodeId)
	nw := kdmlib.InitializeNetwork(3, 12000, rt, nodeId, false)
	kdmlib.NewKademliaInstance(nw, nodeId, kdmlib.ALPHA, kdmlib.K, rt)
}
