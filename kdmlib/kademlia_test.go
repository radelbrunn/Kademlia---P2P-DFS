package kdmlib

import (
	"fmt"
	"testing"
	"time"
)

func TestKademlia_SortContacts(t *testing.T) {
	nodeId := "11100000"
	rt := CreateAllWorkersForRoutingTable(K, IDLENGTH, 5, nodeId)
	nw := InitNetwork("12000", "127.0.0.1", rt, nodeId, true)
	testKademlia := NewKademliaInstance(nw, nodeId, ALPHA, K, rt)

	target := "11111111"

	testContacts := []AddressTriple{
		{"0", "00", "10010000"}, {"0", "00", "00000001"},
		{"0", "00", "00000011"}, {"0", "00", "00000100"},
		{"0", "00", "01100000"}, {"0", "00", "11110000"},
		{"0", "00", "11010000"}, {"0", "00", "11111101"},
		{"0", "00", "01110000"}, {"0", "00", "11110001"},
		{"0", "00", "11110010"}, {"0", "00", "11110011"},
		{"0", "00", "11110100"}, {"0", "00", "11110101"},
		{"0", "00", "11110110"}, {"0", "00", "11110111"},
		{"0", "00", "11111000"}, {"0", "00", "11111001"},
		{"0", "00", "11111010"}, {"0", "00", "11111011"},
		{"0", "00", "11111100"}, {"0", "00", "10110000"}}

	for _, e := range testContacts {
		testKademlia.closest = append(testKademlia.closest, e)
	}

	testKademlia.sortContacts(target)

	if testKademlia.closest[0].Id != "11111101" || testKademlia.closest[len(testKademlia.closest)-1].Id != "00000100" {
		t.Error("Sorting is all wrong")
		t.Fail()
	}

	if len(testKademlia.closest) > testKademlia.k {
		t.Error("List of closest is too long (expected at most ", testKademlia.k)
		t.Fail()
	}

	if len(testContacts) > 20 {
		if len(testKademlia.closest) != 20 {
			t.Error("List of closest does not have", testKademlia.k, " elements")
			t.Fail()
		}
	}
}

func TestKademlia_RefreshClosest(t *testing.T) {
	nodeId := "11100000"
	rt := CreateAllWorkersForRoutingTable(K, IDLENGTH, 5, nodeId)
	nw := InitNetwork("12000", "127.0.0.1", rt, nodeId, true)
	testKademlia := NewKademliaInstance(nw, nodeId, ALPHA, K, rt)

	target := "11111111"

	testContacts := []AddressTriple{
		{"0", "00", "10010000"}, {"0", "00", "00000001"},
		{"0", "00", "00000011"}, {"0", "00", "00000100"},
		{"0", "00", "01100000"}, {"0", "00", "11110000"},
		{"0", "00", "11010000"}, {"0", "00", "11111101"},
		{"0", "00", "01110000"}, {"0", "00", "11110001"},
		{"0", "00", "11110010"}, {"0", "00", "11110011"},
		{"0", "00", "11110100"}, {"0", "00", "11110101"},
		{"0", "00", "11110110"}, {"0", "00", "11110111"},
		{"0", "00", "11111000"}, {"0", "00", "11111001"},
		{"0", "00", "11111010"}, {"0", "00", "11111011"},
		{"0", "00", "11111100"}, {"0", "00", "10110000"}}

	for _, e := range testContacts[0:6] {
		testKademlia.closest = append(testKademlia.closest, e)
	}

	testKademlia.sortContacts(target)
	intermediate := testKademlia.closest

	newContacts := testContacts[0:6]
	testKademlia.refreshClosest(newContacts, target)

	if len(testKademlia.closest) != len(intermediate) {
		t.Error("Expected same length")
		t.Fail()
	}

	for index := range testKademlia.closest {
		if testKademlia.closest[index].Id != intermediate[index].Id {
			t.Error("Difference in slices")
			t.Fail()
		}
	}

	newContacts = testContacts[6:9]
	testKademlia.refreshClosest(newContacts, target)

	if len(testKademlia.closest) != 9 {
		t.Error("Expected length 9, got", len(testKademlia.closest))
		t.Fail()
	}
}

func TestKademlia_GetNextContact_AskedAllContacts(t *testing.T) {
	nodeId := "11100000"
	rt := CreateAllWorkersForRoutingTable(K, IDLENGTH, 5, nodeId)
	nw := InitNetwork("12000", "127.0.0.1", rt, nodeId, true)
	testKademlia := NewKademliaInstance(nw, nodeId, ALPHA, K, rt)

	t1 := AddressTriple{"t1", "00", "10010000"}
	t2 := AddressTriple{"t2", "00", "01000000"}

	testKademlia.closest = append(testKademlia.closest, t1)
	testKademlia.askedClosest = append(testKademlia.askedClosest, t1)
	testKademlia.gotResultBack = append(testKademlia.gotResultBack, t1)

	nextNode := testKademlia.getNextContact()
	askedAll := testKademlia.askedAllContacts()

	if nextNode != nil {
		t.Error("Expected nil, got ", nextNode)
		t.Fail()
	}
	if askedAll == false {
		t.Error("Asked all returns false, despite having asked all contacts so far")
		t.Fail()
	}

	testKademlia.closest = append(testKademlia.closest, t2)
	nextNode = testKademlia.getNextContact()

	askedAll = testKademlia.askedAllContacts()

	if nextNode.Id != t2.Id {
		t.Error("Expected", t2, " got ", nextNode)
	}
	if askedAll == true {
		t.Error("Asked all returned true, despite not having asked all contacts so far")
	}

	testKademlia.askedClosest = append(testKademlia.askedClosest, t2)
	nextNode = testKademlia.getNextContact()
	askedAll = testKademlia.askedAllContacts()

	if nextNode != nil {
		t.Error("Expected nil, got ", nextNode)
		t.Fail()
	}
	if askedAll == true {
		t.Error("Asked all returned true")
		t.Fail()
	}
}

func TestKademlia_LookupContactTimeout(t *testing.T) {
	nodeId := "11100000"
	targetContact := AddressTriple{"127.0.0.17", "11000", "11111111"}

	testContacts := []AddressTriple{
		{"127.0.0.11", "11100", "00010000"}, {"127.0.0.12", "11100", "00100000"},
		{"127.0.0.13", "11100", "01000000"}, {"127.0.0.14", "11100", "10000000"},
		{"127.0.0.15", "11100", "00110000"}, {"127.0.0.16", "11100", "11110000"}}

	rt := CreateAllWorkersForRoutingTable(K, 8, 5, nodeId)

	for _, e := range testContacts {
		rt.GiveOrder(OrderForRoutingTable{ADD, e, false})
	}

	time.Sleep(time.Second * 2)

	nw := InitNetwork("12000", "127.0.0.1", rt, nodeId, true)
	testKademlia := NewKademliaInstance(nw, nodeId, ALPHA, K, rt)

	contacts, file := testKademlia.LookupAlgorithm(targetContact.Id, ContactLookup)

	if file != nil {
		t.Error("Did not expect a data return")
		t.Fail()
	}

	if len(contacts) != len(testContacts) {
		t.Error("Did not expect addition of more contacts (as all requests timedOut")
		t.Fail()
	}

	if len(testKademlia.rt.FindKClosest(targetContact.Id)) != 0 && len(rt.FindKClosest(targetContact.Id)) != 0 {
		t.Error("Routing tables are inconsistent")
		t.Fail()
	}
}

func TestKademlia_LookupDataTimeout(t *testing.T) {
	nodeId := "11100000"
	targetData := "11111111"

	testContacts := []AddressTriple{
		{"127.0.0.11", "11100", "00010000"}, {"127.0.0.12", "11100", "00100000"},
		{"127.0.0.13", "11100", "01000000"}, {"127.0.0.14", "11100", "10000000"},
		{"127.0.0.15", "11100", "00110000"}, {"127.0.0.16", "11100", "11110000"}}

	rt := CreateAllWorkersForRoutingTable(K, 8, 5, nodeId)

	for _, e := range testContacts {
		rt.GiveOrder(OrderForRoutingTable{ADD, e, false})
	}

	time.Sleep(time.Second * 2)

	nw := InitNetwork("12000", "127.0.0.1", rt, nodeId, true)
	testKademlia := NewKademliaInstance(nw, nodeId, ALPHA, K, rt)

	dataReturned := testKademlia.LookupData(targetData, true)

	if dataReturned == true {
		t.Error("Did not expect a data return")
		t.Fail()
	}

	if len(testKademlia.rt.FindKClosest(targetData)) != 0 && len(rt.FindKClosest(targetData)) != 0 {
		t.Error("Routing tables are inconsistent")
		t.Fail()
	}
}

func TestKademlia_LookupEmptyRT(t *testing.T) {
	nodeId := "11100000"
	targetContact := AddressTriple{"127.0.0.17", "11000", "11111111"}

	rt := CreateAllWorkersForRoutingTable(K, 8, 5, nodeId)

	time.Sleep(time.Second * 2)

	nw := InitNetwork("12000", "127.0.0.1", rt, nodeId, true)
	testKademlia := NewKademliaInstance(nw, nodeId, ALPHA, K, rt)

	contacts, file := testKademlia.LookupAlgorithm(targetContact.Id, ContactLookup)

	if contacts != nil || file != nil {
		fmt.Println("Expected nil returns, as the routing table was initially empty")
	}
}
