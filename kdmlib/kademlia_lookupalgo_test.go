package kdmlib

import (
	"Kademlia---P2P-DFS/kdmlib/fileutils"
	"math/rand"
	"testing"
	"time"
)

func TestKademlia_LookupAlgorithm(t *testing.T) {
	nodeId1 := "10010000"
	port1 := "12000"
	nodeId2 := "10110000"
	port2 := "13000"
	nodeId3 := "00000001"
	port3 := "12001"

	targetContact := AddressTriple{"127.0.0.1", "14000", "11111111"}

	testContacts := []AddressTriple{
		{"127.0.0.1", "12000", "10010000"}, {"127.0.0.1", "12001", "00000001"},
		{"127.0.0.1", "12002", "00000011"}, {"127.0.0.1", "12003", "00000100"},
		{"127.0.0.1", "12004", "01100000"}, {"127.0.0.1", "12005", "11110000"},
		{"127.0.0.1", "12006", "11010000"}, {"127.0.0.1", "12007", "11111101"},
		{"127.0.0.1", "12008", "01110000"}, {"127.0.0.1", "12009", "11110001"},
		{"127.0.0.1", "12010", "11110010"}, {"127.0.0.1", "12011", "11110011"},
		{"127.0.0.1", "12012", "11110100"}, {"127.0.0.1", "12013", "11110101"},
		{"127.0.0.1", "12014", "11110110"}, {"127.0.0.1", "12015", "11110111"},
		{"127.0.0.1", "12016", "11111000"}, {"127.0.0.1", "12017", "11111001"},
		{"127.0.0.1", "12018", "11111010"}, {"127.0.0.1", "12019", "11111011"},
		{"127.0.0.1", "12020", "11111100"}, {"127.0.0.1", "13000", "10110000"}}

	rt1 := CreateAllWorkersForRoutingTable(K, 8, 5, nodeId1)
	for _, e := range testContacts[1:11] {
		rt1.GiveOrder(OrderForRoutingTable{ADD, e, false})
	}

	time.Sleep(time.Second * 1)

	rt2 := CreateAllWorkersForRoutingTable(K, 8, 5, nodeId2)
	for _, e := range testContacts[0:1] {
		rt2.GiveOrder(OrderForRoutingTable{ADD, e, false})
	}

	time.Sleep(time.Second * 1)

	rt3 := CreateAllWorkersForRoutingTable(K, 8, 5, nodeId2)
	for _, e := range testContacts[11:] {
		rt3.GiveOrder(OrderForRoutingTable{ADD, e, false})
	}

	time.Sleep(time.Second * 1)

	chanPin1, chanFile1, fileMap1 := fileUtilsKademlia.CreateAndLaunchFileWorkers()
	chanPin2, chanFile2, fileMap2 := fileUtilsKademlia.CreateAndLaunchFileWorkers()
	chanPin3, chanFile3, fileMap3 := fileUtilsKademlia.CreateAndLaunchFileWorkers()

	InitNetwork(port1, "127.0.0.1", rt1, nodeId1, false, true, chanFile1, chanPin1, fileMap1)
	InitNetwork(port3, "127.0.0.1", rt3, nodeId3, false, true, chanFile3, chanPin3, fileMap3)
	nw2 := InitNetwork(port2, "127.0.0.1", rt2, nodeId2, true, true, chanFile2, chanPin2, fileMap2)

	testKademlia := NewKademliaInstance(nw2, nodeId2, ALPHA, K, rt2, chanFile2, fileMap2)

	contacts, contactWithData := testKademlia.LookupAlgorithm(targetContact.Id, ContactLookup)

	testKademlia.closest = []AddressTriple{}
	for _, e := range testContacts {
		testKademlia.closest = append(testKademlia.closest, e)
	}

	testKademlia.sortContacts(targetContact.Id)

	for i, e := range testKademlia.closest {
		if e != contacts[i] {
			t.Error("Difference in slices")
			t.Fail()
		}
	}

	if contactWithData.Id != "" {
		t.Error("Did not expect a data return")
		t.Fail()
	}
}

func TestKademlia_LookupAlgorithm_160(t *testing.T) {

	time.Sleep(time.Second * 2)

	testContacts := []AddressTriple{
		{"127.0.0.1", "22000", GenerateRandID(int64(rand.Intn(100)), 160)}, {"127.0.0.1", "22001", GenerateRandID(int64(rand.Intn(100)), 160)},
		{"127.0.0.1", "22002", GenerateRandID(int64(rand.Intn(100)), 160)}, {"127.0.0.1", "22003", GenerateRandID(int64(rand.Intn(100)), 160)},
		{"127.0.0.1", "22004", GenerateRandID(int64(rand.Intn(100)), 160)}, {"127.0.0.1", "22005", GenerateRandID(int64(rand.Intn(100)), 160)},
		{"127.0.0.1", "22006", GenerateRandID(int64(rand.Intn(100)), 160)}, {"127.0.0.1", "22007", GenerateRandID(int64(rand.Intn(100)), 160)}}

	nodeId1 := testContacts[0].Id
	port1 := testContacts[0].Port
	nodeId2 := testContacts[4].Id
	port2 := testContacts[4].Port
	nodeId3 := testContacts[6].Id
	port3 := testContacts[6].Port

	targetContact := testContacts[7]

	rt1 := CreateAllWorkersForRoutingTable(K, 160, 5, nodeId1)
	for _, e := range testContacts[1:5] {
		rt1.GiveOrder(OrderForRoutingTable{ADD, e, false})
	}

	time.Sleep(time.Second * 1)

	rt2 := CreateAllWorkersForRoutingTable(K, 160, 5, nodeId2)
	for _, e := range testContacts[2:7] {
		rt2.GiveOrder(OrderForRoutingTable{ADD, e, false})
	}

	time.Sleep(time.Second * 1)

	rt3 := CreateAllWorkersForRoutingTable(K, 160, 5, nodeId2)
	for _, e := range testContacts[2:] {
		rt3.GiveOrder(OrderForRoutingTable{ADD, e, false})
	}

	time.Sleep(time.Second * 1)

	chanPin1, chanFile1, fileMap1 := fileUtilsKademlia.CreateAndLaunchFileWorkers()
	chanPin2, chanFile2, fileMap2 := fileUtilsKademlia.CreateAndLaunchFileWorkers()
	chanPin3, chanFile3, fileMap3 := fileUtilsKademlia.CreateAndLaunchFileWorkers()

	nw1 := InitNetwork(port1, "127.0.0.1", rt1, nodeId1, false, true, chanFile1, chanPin1, fileMap1)
	InitNetwork(port2, "127.0.0.1", rt2, nodeId2, false, true, chanFile2, chanPin2, fileMap2)
	InitNetwork(port3, "127.0.0.1", rt3, nodeId3, false, true, chanFile3, chanPin3, fileMap3)

	testKademlia := NewKademliaInstance(nw1, nodeId1, ALPHA, K, rt1, chanFile1, fileMap1)

	contacts, contactWithData := testKademlia.LookupAlgorithm(targetContact.Id, ContactLookup)

	if len(contacts) != 7 {
		t.Error("Expected to get 7 contacts back, got", len(contacts))
		t.Fail()
	}

	if contactWithData.Id != "" {
		t.Error("Did not expect a data return")
		t.Fail()
	}
}
