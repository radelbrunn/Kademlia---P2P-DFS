package main

import (
	"Kademlia---P2P-DFS/kdmlib"
	"fmt"
	"math/rand"
	"net"
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

	answerChannel := make(chan interface{})
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12000")

	nw := kdmlib.InitializeNetwork(5, 12000, routingT, false)
	nw2 := kdmlib.InitializeNetwork(5, 9000, routingT, false)
	nw2.SendPing(addr, answerChannel)
	fmt.Println(nw)

}

func StartKademlia() {
	nodeId := kdmlib.GenerateRandID(int64(rand.Intn(100)))
	rt := kdmlib.CreateAllWorkersForRoutingTable(kdmlib.K, kdmlib.IDLENGTH, 5, nodeId)
	nw := kdmlib.InitializeNetwork(3, 12000, rt, false)
	kdmlib.NewKademliaInstance(nw, nodeId, kdmlib.ALPHA, kdmlib.K, rt)
}
