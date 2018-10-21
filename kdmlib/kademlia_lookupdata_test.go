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

func TestKademlia_LookupDataFail(t *testing.T) {

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
	chanPin1, chanFile1, fileMap1 := fileUtilsKademlia.CreateAndLaunchFileWorkers()
	chanPin2, chanFile2, fileMap2 := fileUtilsKademlia.CreateAndLaunchFileWorkers()
	chanPin3, chanFile3, fileMap3 := fileUtilsKademlia.CreateAndLaunchFileWorkers()

	InitNetwork(port1, "127.0.0.1", rt1, nodeId1, false, true, chanFile1, chanPin1, fileMap1)
	InitNetwork(port3, "127.0.0.1", rt3, nodeId3, false, true, chanFile3, chanPin3, fileMap3)
	nw2 := InitNetwork(port2, "127.0.0.1", rt2, nodeId2, true, true, chanFile2, chanPin2, fileMap2)

	testKademlia := NewKademliaInstance(nw2, nodeId2, ALPHA, K, rt2, chanFile2, fileMap2)

	dataReturn := testKademlia.LookupData(targetData)

	closest := testKademlia.closest

	testKademlia.sortContacts(targetData)

	for i, e := range testKademlia.closest {
		if e != closest[i] {
			t.Error("Difference in slices")
			t.Fail()
		}
	}

	if dataReturn != nil {
		t.Error("Expected data return to fail")
		t.Fail()
	}
}

func TestKademlia_LookupDataSuccess(t *testing.T) {

	fileName := "11111111"
	ioutil.WriteFile(fileUtilsKademlia.FileDirectory+ConvertToHexAddr(fileName), []byte("hello world"), 0644)

	nodeId1 := "11111001"
	port1 := "12017"
	nodeId2 := "00000100"
	port2 := "12003"

	testContacts := []AddressTriple{
		{"127.0.0.1", "12003", "00000100"}, {"127.0.0.1", "12002", "00000011"},
		{"127.0.0.1", "12004", "01100000"}, {"127.0.0.1", "12005", "11110000"},
		{"127.0.0.1", "12006", "11010000"}, {"127.0.0.1", "12007", "11111101"},
		{"127.0.0.1", "12008", "01110000"}, {"127.0.0.1", "12009", "11110001"},
		{"127.0.0.1", "12010", "11110010"}, {"127.0.0.1", "12011", "11110011"},
		{"127.0.0.1", "12012", "11110100"}, {"127.0.0.1", "12013", "11110101"},
		{"127.0.0.1", "12014", "11110110"}, {"127.0.0.1", "12015", "11110111"},
		{"127.0.0.1", "12016", "11111000"}, {"127.0.0.1", "12017", "11111001"},
		{"127.0.0.1", "12018", "11111010"}, {"127.0.0.1", "12019", "11111011"},
		{"127.0.0.1", "12020", "11111100"}}

	rt1 := CreateAllWorkersForRoutingTable(K, 8, 5, nodeId1)
	for _, e := range testContacts[0:11] {
		rt1.GiveOrder(OrderForRoutingTable{ADD, e, false})
	}

	time.Sleep(time.Second * 1)

	rt2 := CreateAllWorkersForRoutingTable(K, 8, 5, nodeId2)
	for _, e := range testContacts[1:] {
		rt2.GiveOrder(OrderForRoutingTable{ADD, e, false})
	}

	time.Sleep(time.Second * 1)

	chanPin1, chanFile1, fileMap1 := fileUtilsKademlia.CreateAndLaunchFileWorkers()
	chanPin2, chanFile2, fileMap2 := fileUtilsKademlia.CreateAndLaunchFileWorkers()

	InitNetwork(port1, "127.0.0.1", rt1, nodeId1, false, true, chanFile1, chanPin1, fileMap1)
	nw2 := InitNetwork(port2, "127.0.0.1", rt2, nodeId2, true, true, chanFile2, chanPin2, fileMap2)

	testKademlia := NewKademliaInstance(nw2, nodeId2, ALPHA, K, rt2, chanFile2, fileMap2)

	dataReturn := testKademlia.LookupData(fileName)

	fmt.Println("Data returned: ", string(dataReturn))

	if dataReturn == nil {
		t.Error("Expected to get data back...")
		t.Fail()
	}

	if string(dataReturn) != "hello world" {
		t.Error("Expected to get data back 'hello world, got", string(dataReturn))
		t.Fail()
	}

	os.Remove(fileUtilsKademlia.FileDirectory + ConvertToHexAddr(fileName))
}

func TestKademlia_LookupData_160(t *testing.T) {

	fileName := "1011000010000101111101010010000000001011110101000000001000101001001111010110111001001010001000110110001110001010001000010011001100011001010000001110110011010011"
	ioutil.WriteFile(fileUtilsKademlia.FileDirectory+ConvertToHexAddr(fileName), []byte("hello world"), 0644)

	time.Sleep(time.Second * 2)

	testContacts := []AddressTriple{
		{"127.0.0.1", "22000", GenerateRandID(int64(rand.Intn(100)), 160)}, {"127.0.0.1", "22001", GenerateRandID(int64(rand.Intn(100)), 160)},
		{"127.0.0.1", "22002", GenerateRandID(int64(rand.Intn(100)), 160)}, {"127.0.0.1", "22003", GenerateRandID(int64(rand.Intn(100)), 160)},
		{"127.0.0.1", "22004", GenerateRandID(int64(rand.Intn(100)), 160)}, {"127.0.0.1", "22005", GenerateRandID(int64(rand.Intn(100)), 160)},
		{"127.0.0.1", "22006", GenerateRandID(int64(rand.Intn(100)), 160)}, {"127.0.0.1", "22007", GenerateRandID(int64(rand.Intn(100)), 160)}}

	nodeId1 := testContacts[0].Id
	port1 := testContacts[0].Port
	nodeId2 := testContacts[5].Id
	port2 := testContacts[5].Port

	rt1 := CreateAllWorkersForRoutingTable(K, 160, 5, nodeId1)
	for _, e := range testContacts[1:8] {
		rt1.GiveOrder(OrderForRoutingTable{ADD, e, false})
	}

	time.Sleep(time.Second * 2)

	rt2 := CreateAllWorkersForRoutingTable(K, 160, 5, nodeId2)
	for _, e := range testContacts[2:7] {
		rt2.GiveOrder(OrderForRoutingTable{ADD, e, false})
	}

	time.Sleep(time.Second * 2)

	chanPin1, chanFile1, fileMap1 := fileUtilsKademlia.CreateAndLaunchFileWorkers()
	chanPin2, chanFile2, fileMap2 := fileUtilsKademlia.CreateAndLaunchFileWorkers()

	nw1 := InitNetwork(port1, "127.0.0.1", rt1, nodeId1, false, true, chanFile1, chanPin1, fileMap1)
	InitNetwork(port2, "127.0.0.1", rt2, nodeId2, false, true, chanFile2, chanPin2, fileMap2)

	testKademlia := NewKademliaInstance(nw1, nodeId1, ALPHA, K, rt1, chanFile1, fileMap1)

	dataReturn := testKademlia.LookupData(fileName)

	fmt.Println("Data returned: ", string(dataReturn))

	if string(dataReturn) != "hello world" {
		t.Error("Expected to get data back 'hello world, got", string(dataReturn))
		t.Fail()
	}

	os.Remove(fileUtilsKademlia.FileDirectory + ConvertToHexAddr(fileName))
}
