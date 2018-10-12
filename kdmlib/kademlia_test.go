package kdmlib

import (
	"testing"
)

func TestKademlia_SortContacts(t *testing.T) {
	nodeId := "11100000"
	rt := CreateAllWorkersForRoutingTable(K, IDLENGTH, 5, nodeId)
	nw := InitNetwork("12000", rt, nodeId, true)
	testKademlia := NewKademliaInstance(nw, nodeId, ALPHA, K, rt)

	target := "11111111"

	testContacts := []AddressTriple{
		{"0", "00", "10010000"}, {"0", "00", "00000001"},
		{"0", "00", "00000011"}, {"0", "00", "00000100"},
		{"0", "00", "01100000"}, {"0", "00", "11110000"},
		{"0", "00", "11110000"}, {"0", "00", "11110000"},
		{"0", "00", "11110000"}, {"0", "00", "11110001"},
		{"0", "00", "11110010"}, {"0", "00", "11110011"},
		{"0", "00", "11110100"}, {"0", "00", "11110101"},
		{"0", "00", "11110110"}, {"0", "00", "11110111"},
		{"0", "00", "11111000"}, {"0", "00", "11111001"},
		{"0", "00", "11111010"}, {"0", "00", "11111011"},
		{"0", "00", "11111100"}, {"0", "00", "11111101"}}

	for _, e := range testContacts {
		testKademlia.closest = append(testKademlia.closest, e)
	}

	testKademlia.sortContacts(target)

	if testKademlia.closest[0].Id != "11111101" || testKademlia.closest[len(testKademlia.closest)-1].Id != "00000100" {
		t.Error("Sorting is all wrong")
	}

	if len(testKademlia.closest) > testKademlia.k {
		t.Error("List of closest is too long (expected at most ", testKademlia.k)
	}

	if len(testContacts) > 20 {
		if len(testKademlia.closest) != 20 {
			t.Error("List of closest does not have", testKademlia.k, " elements")
		}
	}
}

func TestKademlia_RefreshClosest(t *testing.T) {
	nodeId := "1110"
	rt := CreateAllWorkersForRoutingTable(K, IDLENGTH, 5, nodeId)
	nw := InitNetwork("12000", rt, nodeId, true)
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
	testKademlia.sortContacts(target)
	intermediate := testKademlia.closest

	newContacts := []AddressTriple{t1, t2, t3, t4, t5, t6}
	testKademlia.refreshClosest(newContacts, target)

	if len(testKademlia.closest) != len(intermediate) {
		t.Error("Expected same length")
	}

	for index := range testKademlia.closest {
		if testKademlia.closest[index].Id != intermediate[index].Id {
			t.Error("Difference in slices")
		}
	}

	newContacts = []AddressTriple{t7, t8, t9}
	testKademlia.refreshClosest(newContacts, target)

	if len(testKademlia.closest) != 9 {
		t.Error("Expected length 9, got", len(testKademlia.closest))
	}
}

func TestKademlia_GetNextContact(t *testing.T) {
	nodeId := "1110"
	rt := CreateAllWorkersForRoutingTable(K, IDLENGTH, 5, nodeId)
	nw := InitNetwork("12000", rt, nodeId, true)
	testKademlia := NewKademliaInstance(nw, nodeId, ALPHA, K, rt)

	t1 := AddressTriple{"t1", "00", "1001"}
	t2 := AddressTriple{"t2", "00", "0100"}

	testKademlia.closest = append(testKademlia.closest, t1)
	testKademlia.askedClosest = append(testKademlia.askedClosest, t1)

	nextNode := testKademlia.getNextContact()
	if nextNode != nil {
		t.Error("Expected nil, got ", nextNode)
	}

	testKademlia.closest = append(testKademlia.closest, t2)

	nextNode = testKademlia.getNextContact()
	if nextNode.Id != t2.Id {
		t.Error("Expected", t2, " got ", nextNode)
	}

	testKademlia.askedClosest = append(testKademlia.askedClosest, t2)
	nextNode = testKademlia.getNextContact()
	if nextNode != nil {
		t.Error("Expected nil, got ", nextNode)
	}
}
