package kdmlib

import (
	"Kademlia---P2P-DFS/kdmlib/fileutils"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestKademlia_StoreData(t *testing.T) {
	fileName := "11111111"

	ioutil.WriteFile(fileUtilsKademlia.FileDirectory+fileName, []byte("hello world"), 0644)

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

	fmt.Println("Populating Routing Table of node: '" + ConvertToHexAddr(nodeId1) + "'...")

	rt1 := CreateAllWorkersForRoutingTable(K, 8, 5, nodeId1)
	for _, e := range testContacts[1:11] {
		rt1.GiveOrder(OrderForRoutingTable{ADD, e, false})
	}

	time.Sleep(time.Millisecond * 1500)

	fmt.Println("Populating Routing Table of node: '" + ConvertToHexAddr(nodeId2) + "'...")

	rt2 := CreateAllWorkersForRoutingTable(K, 8, 5, nodeId2)
	for _, e := range testContacts[0:1] {
		rt2.GiveOrder(OrderForRoutingTable{ADD, e, false})
	}

	time.Sleep(time.Millisecond * 1500)

	fmt.Println("Populating Routing Table of node: '" + ConvertToHexAddr(nodeId3) + "'...")

	rt3 := CreateAllWorkersForRoutingTable(K, 8, 5, nodeId3)
	for _, e := range testContacts[11:21] {
		rt3.GiveOrder(OrderForRoutingTable{ADD, e, false})
	}

	time.Sleep(time.Millisecond * 1500)

	fmt.Println("Populating Routing Table of node: '" + ConvertToHexAddr(nodeId4) + "'...")

	rt4 := CreateAllWorkersForRoutingTable(K, 8, 5, nodeId4)
	for _, e := range testContacts[11:14] {
		rt4.GiveOrder(OrderForRoutingTable{ADD, e, false})
	}

	time.Sleep(time.Millisecond * 1500)

	chanPin1, chanFile1, fileMap1 := fileUtilsKademlia.CreateAndLaunchFileWorkers()
	chanPin2, chanFile2, fileMap2 := fileUtilsKademlia.CreateAndLaunchFileWorkers()
	chanPin3, chanFile3, fileMap3 := fileUtilsKademlia.CreateAndLaunchFileWorkers()
	chanPin4, chanFile4, fileMap4 := fileUtilsKademlia.CreateAndLaunchFileWorkers()

	InitNetwork(port1, "127.0.0.1", rt1, nodeId1, false, true, chanFile1, chanPin1, fileMap1)
	InitNetwork(port3, "127.0.0.1", rt3, nodeId3, false, true, chanFile3, chanPin3, fileMap3)
	InitNetwork(port4, "127.0.0.1", rt4, nodeId4, false, true, chanFile4, chanPin4, fileMap4)
	nw2 := InitNetwork(port2, "127.0.0.1", rt2, nodeId2, false, true, chanFile2, chanPin2, fileMap2)

	testKademlia := NewKademliaInstance(nw2, nodeId2, ALPHA, K, rt2, chanFile2, fileMap2)

	testKademlia.StoreData(fileName)

	time.Sleep(time.Millisecond * 1000)

	data := fileUtilsKademlia.ReadFileFromOS(nodeId1)

	if string(data) != "hello world" {
		t.Error("File was not uploaded")
		t.Fail()
	} else {
		os.Remove(fileUtilsKademlia.FileDirectory + nodeId1)
	}

	data = fileUtilsKademlia.ReadFileFromOS(nodeId4)

	if string(data) != "hello world" {
		t.Error("File was not uploaded")
		t.Fail()
	} else {
		os.Remove(fileUtilsKademlia.FileDirectory + nodeId4)
	}

	os.Remove(fileUtilsKademlia.FileDirectory + fileName)
}

func TestKademlia_StoreData_160(t *testing.T) {
	fileName := "1101011011110000010110101111010011010100001000010110001011000010010100111011111001001100000010101111000101101010111100110000000010110001001000110010110011000111"

	ioutil.WriteFile(fileUtilsKademlia.FileDirectory+fileName, []byte("hello world"), 0644)

	testContacts := []AddressTriple{
		{"127.0.0.1", "24000", GenerateRandID(int64(rand.Intn(100)), 160)}, {"127.0.0.1", "24001", GenerateRandID(int64(rand.Intn(100)), 160)},
		{"127.0.0.1", "24002", GenerateRandID(int64(rand.Intn(100)), 160)}, {"127.0.0.1", "24003", GenerateRandID(int64(rand.Intn(100)), 160)},
		{"127.0.0.1", "24004", GenerateRandID(int64(rand.Intn(100)), 160)}, {"127.0.0.1", "24005", GenerateRandID(int64(rand.Intn(100)), 160)},
		{"127.0.0.1", "24006", GenerateRandID(int64(rand.Intn(100)), 160)}, {"127.0.0.1", "24007", GenerateRandID(int64(rand.Intn(100)), 160)}}

	nodeId1 := testContacts[0].Id
	port1 := testContacts[0].Port
	nodeId2 := testContacts[3].Id
	port2 := testContacts[3].Port
	nodeId3 := testContacts[5].Id
	port3 := testContacts[5].Port

	fmt.Println("Populating Routing Table of node: '" + ConvertToHexAddr(nodeId1) + "'...")

	rt1 := CreateAllWorkersForRoutingTable(K, 160, 5, nodeId1)
	for _, e := range testContacts[1:] {
		rt1.GiveOrder(OrderForRoutingTable{ADD, e, false})
	}

	time.Sleep(time.Millisecond * 2500)

	fmt.Println("Populating Routing Table of node: '" + ConvertToHexAddr(nodeId2) + "'...")

	rt2 := CreateAllWorkersForRoutingTable(K, 160, 5, nodeId2)
	for _, e := range testContacts[2:6] {
		rt2.GiveOrder(OrderForRoutingTable{ADD, e, false})
	}

	time.Sleep(time.Millisecond * 2500)

	fmt.Println("Populating Routing Table of node: '" + ConvertToHexAddr(nodeId1) + "'...")

	rt3 := CreateAllWorkersForRoutingTable(K, 160, 5, nodeId3)
	for _, e := range testContacts[0:3] {
		rt3.GiveOrder(OrderForRoutingTable{ADD, e, false})
	}

	time.Sleep(time.Millisecond * 2500)

	chanPin1, chanFile1, fileMap1 := fileUtilsKademlia.CreateAndLaunchFileWorkers()
	chanPin2, chanFile2, fileMap2 := fileUtilsKademlia.CreateAndLaunchFileWorkers()
	chanPin3, chanFile3, fileMap3 := fileUtilsKademlia.CreateAndLaunchFileWorkers()

	nw1 := InitNetwork(port1, "127.0.0.1", rt1, nodeId1, false, true, chanFile1, chanPin1, fileMap1)
	InitNetwork(port2, "127.0.0.1", rt2, nodeId2, false, true, chanFile2, chanPin2, fileMap2)
	InitNetwork(port3, "127.0.0.1", rt3, nodeId3, false, true, chanFile3, chanPin3, fileMap3)

	testKademlia := NewKademliaInstance(nw1, nodeId1, ALPHA, K, rt1, chanFile1, fileMap1)

	testKademlia.StoreData(fileName)

	time.Sleep(time.Second * 1)

	data := fileUtilsKademlia.ReadFileFromOS(nodeId2)

	if string(data) != "hello world" {
		t.Error("File was not uploaded")
		t.Fail()
	} else {
		os.Remove(fileUtilsKademlia.FileDirectory + nodeId2)
	}

	data = fileUtilsKademlia.ReadFileFromOS(nodeId3)

	if string(data) != "hello world" {
		t.Error("File was not uploaded")
		t.Fail()
	} else {
		os.Remove(fileUtilsKademlia.FileDirectory + nodeId3)
	}

	os.Remove(fileUtilsKademlia.FileDirectory + ConvertToHexAddr(fileName))
}
