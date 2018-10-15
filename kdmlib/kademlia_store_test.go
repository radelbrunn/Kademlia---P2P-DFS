package kdmlib

import (
	"Kademlia---P2P-DFS/kdmlib/fileutils"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

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
}
