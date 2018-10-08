package kdmlib

import (
	"math/rand"
	"testing"
)

func TestKademlia_SortContacts(t *testing.T) {
	nodeId := GenerateRandID(int64(rand.Intn(100)))
	rt := CreateAllWorkersForRoutingTable(K, IDLENGTH, 5, nodeId)
	nw := InitializeNetwork(3, 12000, rt, true)
	testKademlia := NewKademliaInstance(nw, nodeId, ALPHA, K, rt)

	target := "1111"

	t1 := AddressTriple{"0", "00", "1001"}
	t2 := AddressTriple{"0", "00", "0100"}
	t3 := AddressTriple{"0", "00", "0010"}
	t4 := AddressTriple{"0", "00", "1000"}
	t5 := AddressTriple{"0", "00", "0110"}
	t6 := AddressTriple{"0", "00", "1111"}

	testKademlia.closest = append(testKademlia.closest, t1, t2, t3, t4, t5, t6)
	testKademlia.SortContacts(target)

	if testKademlia.closest[0] != t6 || testKademlia.closest[len(testKademlia.closest)-1] != t3 {
		t.Error("Sorting is all wrong")
	}
}

func TestKademlia_RefreshClosest(t *testing.T) {
	nodeId := GenerateRandID(int64(rand.Intn(100)))
	rt := CreateAllWorkersForRoutingTable(K, IDLENGTH, 5, nodeId)
	nw := InitializeNetwork(3, 12000, rt, true)
	testKademlia := NewKademliaInstance(nw, nodeId, ALPHA, K, rt)

	target := "1111"

	t1 := AddressTriple{"t1", "00", "1001"}
	t2 := AddressTriple{"t2", "00", "0100"}
	t3 := AddressTriple{"t3", "00", "0010"}
	t4 := AddressTriple{"t4", "00", "1000"}
	t5 := AddressTriple{"t5", "00", "0110"}
	t6 := AddressTriple{"t6", "00", "1111"}
	t7 := AddressTriple{"t7", "00", "0001"}
	t8 := AddressTriple{"t8", "00", "0011"}
	t9 := AddressTriple{"t9", "00", "1011"}

	testKademlia.closest = append(testKademlia.closest, t1, t2, t3, t4, t5, t6)
	testKademlia.SortContacts(target)
	intermediate := testKademlia.closest

	newContacts := []AddressTriple{t1, t2, t3, t4, t5, t6}
	testKademlia.RefreshClosest(newContacts, target)

	if len(testKademlia.closest) != len(intermediate) {
		t.Error("Expected same length")
	}

	for index := range testKademlia.closest {
		if testKademlia.closest[index].Id != intermediate[index].Id {
			t.Error("Difference in slices")
		}
	}

	newContacts = []AddressTriple{t7, t8, t9}
	testKademlia.RefreshClosest(newContacts, target)

	if len(testKademlia.closest) != 9 {
		t.Error("Expected length 9, got", len(testKademlia.closest))
	}
}

func TestKademlia_GetNextContact(t *testing.T) {
	nodeId := GenerateRandID(int64(rand.Intn(100)))
	rt := CreateAllWorkersForRoutingTable(K, IDLENGTH, 5, nodeId)
	nw := InitializeNetwork(3, 12000, rt, true)
	testKademlia := NewKademliaInstance(nw, nodeId, ALPHA, K, rt)

	t1 := AddressTriple{"t1", "00", "1001"}
	t2 := AddressTriple{"t2", "00", "0100"}

	testKademlia.closest = append(testKademlia.closest, t1)
	testKademlia.asked[testKademlia.closest[0].Id] = true

	nextNode := testKademlia.GetNextContact()
	if nextNode != nil {
		t.Error("Expected nil, got ", nextNode)
	}

	testKademlia.closest = append(testKademlia.closest, t2)

	nextNode = testKademlia.GetNextContact()
	if nextNode.Id != t2.Id {
		t.Error("Expected", t2, " got ", nextNode)
	}

	testKademlia.asked[testKademlia.closest[1].Id] = true
	nextNode = testKademlia.GetNextContact()
	if nextNode != nil {
		t.Error("Expected nil, got ", nextNode)
	}
}
