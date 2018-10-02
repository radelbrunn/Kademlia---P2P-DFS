package kdmlib

import (
	"math/rand"
	"testing"
	"time"
)

func TestSortingContacts(t *testing.T) {
	nodeId := GenerateRandID(int64(rand.Intn(100)))
	rt := CreateAllWorkersForRoutingTable(K, IDLENGTH, 5, nodeId)
	nw := InitializeNetwork(3, 12000, rt, true)
	kademlia := NewKademliaInstance(nw, nodeId, ALPHA, K, rt)

	target := "1111"

	t1 := AddressTriple{"0", "00", "1001"}
	t2 := AddressTriple{"0", "00", "0100"}
	t3 := AddressTriple{"0", "00", "0010"}
	t4 := AddressTriple{"0", "00", "1000"}
	t5 := AddressTriple{"0", "00", "0110"}
	t6 := AddressTriple{"0", "00", "1111"}

	kademlia.closest = append(kademlia.closest, t1, t2, t3, t4, t5, t6)
	kademlia.SortContacts(target)

	time.Sleep(time.Second)

	if kademlia.closest[0] != t6 || kademlia.closest[len(kademlia.closest)-1] != t3 {
		t.Error("Sorting is all wrong")
	}
}

/**
func TestUpdateContactsEqual(t *testing.T) {

}

**/
