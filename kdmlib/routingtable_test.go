package kdmlib

import (
	"fmt"
	"testing"
	"time"
)

func TestAddRoutingNotExist(t *testing.T) {
	routingTable := CreateAllWorkersForRoutingTable(1, 2, 1, "00")
	routingTable.GiveOrder(OrderForRoutingTable{ADD, AddressTriple{"0", "0", "01"}, false})
	time.Sleep(time.Second) //need to sleep because of the concurrency
	if (*routingTable.routingtable.routingTable)[1].Len() != 1 {
		t.Error("should be addded")
		fmt.Println((*routingTable.routingtable.routingTable)[1].Len())
		fmt.Println((*routingTable.routingtable.routingTable)[0].Len())
	}
}

func TestAddRoutingIfExist(t *testing.T) {
	routingTable := CreateAllWorkersForRoutingTable(1, 2, 1, "00")
	routingTable.GiveOrder(OrderForRoutingTable{ADD, AddressTriple{"0", "0", "01"}, false})
	routingTable.GiveOrder(OrderForRoutingTable{ADD, AddressTriple{"0", "0", "01"}, false})
	time.Sleep(time.Second) //need to sleep because of the concurrency
	if (*routingTable.routingtable.routingTable)[1].Len() != 1 {
		t.Error("should be 1")
		fmt.Println((*routingTable.routingtable.routingTable)[1].Len())
		fmt.Println((*routingTable.routingtable.routingTable)[0].Len())
	}
}

func TestRemove(t *testing.T) {
	routingTable := CreateAllWorkersForRoutingTable(1, 2, 1, "00")
	routingTable.GiveOrder(OrderForRoutingTable{ADD, AddressTriple{"0", "0", "01"}, false})
	routingTable.GiveOrder(OrderForRoutingTable{REMOVE, AddressTriple{"0", "0", "01"}, false})
	time.Sleep(time.Second) //need to sleep because of the concurrency
	if (*routingTable.routingtable.routingTable)[1].Len() != 0 {
		t.Error("should be 1")
		fmt.Println((*routingTable.routingtable.routingTable)[1].Len())
		fmt.Println((*routingTable.routingtable.routingTable)[0].Len())
	}
}

func TestComputeDistanceNotSameLength(t *testing.T) {
	_, res := ComputeDistance("0100", "111")
	if res == nil {
		t.Error("should return an error")
	}
}

func TestBump(t *testing.T) {
	routingTable := CreateAllWorkersForRoutingTable(2, 2, 1, "00")
	routingTable.GiveOrder(OrderForRoutingTable{ADD, AddressTriple{"0", "0", "10"}, false})
	routingTable.GiveOrder(OrderForRoutingTable{ADD, AddressTriple{"0", "0", "11"}, false})
	routingTable.GiveOrder(OrderForRoutingTable{ADD, AddressTriple{"0", "0", "10"}, false})

	time.Sleep(time.Second) //need to sleep because of the concurrency
	if (*routingTable.routingtable.routingTable)[0].Front().Value.(AddressTriple).Id != "10" {
		t.Error("should be 10 at the first place")

	}
}

func TestRoutingTable_FindKClosest(t *testing.T) {
	routingTable := CreateAllWorkersForRoutingTable(2, 2, 1, "00")
	routingTable.GiveOrder(OrderForRoutingTable{ADD, AddressTriple{"0", "0", "10"}, false})
	routingTable.GiveOrder(OrderForRoutingTable{ADD, AddressTriple{"0", "0", "11"}, false})
	routingTable.GiveOrder(OrderForRoutingTable{ADD, AddressTriple{"0", "0", "10"}, false})
	time.Sleep(time.Second) //need to sleep because of the concurrency
	array := routingTable.FindKClosest("00")
	if len(array) != 2 {
		t.Error("should only return 2 values")
	} else {
		if array[0].Distance != "10" {
			t.Error("not the closest one at the first place")
		}
	}
}

func TestRoutingTableOverLoad(t *testing.T) {
	routingTable := CreateAllWorkersForRoutingTable(2, 3, 1, "000")
	routingTable.GiveOrder(OrderForRoutingTable{ADD, AddressTriple{"0", "0", "100"}, false})
	routingTable.GiveOrder(OrderForRoutingTable{ADD, AddressTriple{"0", "0", "110"}, false})
	routingTable.GiveOrder(OrderForRoutingTable{ADD, AddressTriple{"0", "0", "111"}, false})
	time.Sleep(time.Second * 3)

	if (*routingTable.routingtable.routingTable)[0].Front().Value.(AddressTriple).Id != "111" {
		t.Error("wrong triple in the list")
	}
}
