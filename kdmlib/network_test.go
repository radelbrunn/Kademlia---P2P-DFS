package kdmlib

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestNetwork_SendPing(t *testing.T) {
	id1 := "00001"
	id2 := "00010"
	id3 := "00100"
	id4 := "01000"
	routingT := CreateAllWorkersForRoutingTable(20, 5, 5, "00000")
	//routingT := CreateAllWorkersForRoutingTable(20,160,5,GenerateRandID())
	routingT.GiveOrder(OrderForRoutingTable{ADD, AddressTriple{"127.0.0.1", "9000", id1}, false})
	routingT.GiveOrder(OrderForRoutingTable{ADD, AddressTriple{"127.0.0.1", "9000", id2}, false})
	routingT.GiveOrder(OrderForRoutingTable{ADD, AddressTriple{"127.0.0.1", "9000", id3}, false})
	routingT.GiveOrder(OrderForRoutingTable{ADD, AddressTriple{"127.0.0.1", "9000", id4}, false})
	time.Sleep(time.Second)

	network := InitializeNetwork(5, 12000, routingT)

	contactAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:12000") //<-- try this address when testing!
	CheckError(err)

	answerChannel := make(chan interface{})
	go network.SendPing(contactAddr, answerChannel)
	//go network.SendFindContact(contactAddr,"00000", answerChannel)
	answer := <-answerChannel
	switch answer := answer.(type) {
	case bool: //false
		t.Error("Connection timed out!")
		break
	default:
		fmt.Println(answer)
		break
	}
}
