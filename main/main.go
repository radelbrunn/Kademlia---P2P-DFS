package main

import (
	"Kademlia---P2P-DFS/kdmlib"
	"fmt"
	"math/rand"
)

func main() {
	StartKademlia()
}

func StartKademlia() {
	nodeId := kdmlib.GenerateRandID(int64(rand.Intn(100)))
	rt := kdmlib.CreateAllWorkersForRoutingTable(kdmlib.K, kdmlib.IDLENGTH, 5, nodeId)
	nw := kdmlib.InitNetwork("12000", "127.0.0.1", rt, nodeId, false)
	kdmlib.NewKademliaInstance(nw, nodeId, kdmlib.ALPHA, kdmlib.K, rt)
	fmt.Println("XD")
}
