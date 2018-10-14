package kdmlib

import (
	"Kademlia---P2P-DFS/kdmlib/fileutils"
	"fmt"
	"io/ioutil"
	"os"
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
	if askedAll == false {
		t.Error("Asked all returns false, despite having asked all contacts so far")
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

	InitNetwork(port1, "127.0.0.1", rt1, nodeId1, false)
	InitNetwork(port3, "127.0.0.1", rt3, nodeId3, false)
	nw2 := InitNetwork(port2, "127.0.0.1", rt2, nodeId2, true)

	testKademlia := NewKademliaInstance(nw2, nodeId2, ALPHA, K, rt2)

	contacts, data := testKademlia.LookupAlgorithm(targetContact.Id, ContactLookup)

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

	if data != nil {
		t.Error("Did not expect a data return")
		t.Fail()
	}
}

func TestKademlia_StoreData(t *testing.T) {

	os.Mkdir(".files/", 0755)

	fileName := "kdmtestfile"

	ioutil.WriteFile(".files"+string(os.PathSeparator)+fileName, []byte("hello world"), 0644)

	nodeId1 := "10010000"
	port1 := "12000"
	nodeId2 := "10110000"
	port2 := "13000"
	nodeId3 := "00000001"
	port3 := "12001"
	nodeId4 := "11110100"
	port4 := "12012"

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

	rt3 := CreateAllWorkersForRoutingTable(K, 8, 5, nodeId3)
	for _, e := range testContacts[11:] {
		rt3.GiveOrder(OrderForRoutingTable{ADD, e, false})
	}

	time.Sleep(time.Second * 1)

	rt4 := CreateAllWorkersForRoutingTable(K, 8, 5, nodeId4)
	for _, e := range testContacts[11:14] {
		rt4.GiveOrder(OrderForRoutingTable{ADD, e, false})
	}

	time.Sleep(time.Second * 1)

	InitNetwork(port1, "127.0.0.1", rt1, nodeId1, false)
	InitNetwork(port3, "127.0.0.1", rt3, nodeId3, false)
	InitNetwork(port4, "127.0.0.1", rt4, nodeId4, false)
	nw2 := InitNetwork(port2, "127.0.0.1", rt2, nodeId2, true)

	testKademlia := NewKademliaInstance(nw2, nodeId2, ALPHA, K, rt2)

	testKademlia.StoreData(fileName, true)

	time.Sleep(time.Second * 1)

	data := fileUtilsKademlia.ReadFileFromOS("10010000kdmtestfile")

	if string(data) != "hello world" {
		t.Error("File was not uploaded")
		t.Fail()
	} else {
		os.Remove(".files" + string(os.PathSeparator) + "10010000kdmtestfile")
	}

	data = fileUtilsKademlia.ReadFileFromOS("11110100kdmtestfile")

	if string(data) != "hello world" {
		t.Error("File was not uploaded")
		t.Fail()
	} else {
		os.Remove(".files" + string(os.PathSeparator) + "11110100kdmtestfile")
	}

	os.Remove(".files" + string(os.PathSeparator) + "kdmtestfile")
	os.Remove(".files")
}

/* Check Windows compatibility
func TestKademlia_LookupData(t *testing.T) {
	nodeId1 := "10010000"
	port1 := "12000"
	nodeId2 := "10110000"
	port2 := "13000"
	nodeId3 := "00000001"
	port3 := "12001"

	targetData := "11111111"

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

	InitNetwork(port1, "127.0.0.1", rt1, nodeId1, false)
	InitNetwork(port3, "127.0.0.1", rt3, nodeId3, false)
	nw2 := InitNetwork(port2, "127.0.0.1", rt2, nodeId2, true)

	testKademlia := NewKademliaInstance(nw2, nodeId2, ALPHA, K, rt2)

	dataReturn := testKademlia.LookupData(targetData, true)

	closest := testKademlia.closest

	testKademlia.sortContacts(targetData)

	for i, e := range testKademlia.closest {
		if e != closest[i] {
			t.Error("Difference in slices")
			t.Fail()
		}
	}

	if dataReturn != false {
		t.Error("Expected data return to fail")
		t.Fail()
	}
}
*/
