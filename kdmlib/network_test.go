package kdmlib

import (
	"Kademlia---P2P-DFS/github.com/golang/protobuf/proto"
	pb "Kademlia---P2P-DFS/kdmlib/proto_config"
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"reflect"
	"testing"
)

func TestNetwork_SendPing(t *testing.T) {
	nodeId := GenerateRandID(int64(rand.Intn(100)))
	rt := CreateAllWorkersForRoutingTable(K, IDLENGTH, 5, nodeId)
	network := InitializeNetwork(3, 12345, rt, true)

	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12000")
	conn, _ := net.ListenUDP("udp", addr)

	defer conn.Close()

	buf := make([]byte, 4096)

	channel := make(chan interface{})
	go network.SendPing(addr, channel)

	for {
		n, _, _ := conn.ReadFromUDP(buf)
		container := &pb.Container{}
		proto.Unmarshal(buf[0:n], container)

		if container.REQUEST_TYPE != Request || container.REQUEST_ID != Ping ||
			container.GetRequestPing().ID != "asdasd" {
			t.Error("Didn't receive a ping request! Hmmm....")
			t.Fail()
		}
		return
	}
}
func TestNetwork_SendFindContact(t *testing.T) {
	nodeId := GenerateRandID(int64(rand.Intn(100)))
	rt := CreateAllWorkersForRoutingTable(K, IDLENGTH, 5, nodeId)
	network := InitializeNetwork(3, 12000, rt, true)

	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12000")
	conn, _ := net.ListenUDP("udp", addr)

	defer conn.Close()

	buf := make([]byte, 4096)

	channel := make(chan interface{})
	contactID := "00000"
	go network.SendFindContact(addr, contactID, channel)

	for {
		n, _, _ := conn.ReadFromUDP(buf)
		container := &pb.Container{}
		proto.Unmarshal(buf[0:n], container)

		if container.REQUEST_TYPE != Request || container.REQUEST_ID != FindContact ||
			container.GetRequestContact().ID != contactID { //create an addresstripple instead
			t.Error("Didn't receive a FindContact request! Hmmm....")
			t.Fail()
		}
		return
	}
}
func TestNetwork_SendFindData(t *testing.T) {
	nodeId := GenerateRandID(int64(rand.Intn(100)))
	rt := CreateAllWorkersForRoutingTable(K, IDLENGTH, 5, nodeId)
	network := InitializeNetwork(3, 12000, rt, true)

	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12000")
	conn, _ := net.ListenUDP("udp", addr)

	defer conn.Close()

	buf := make([]byte, 4096)

	channel := make(chan interface{})
	hash := "00000"
	go network.SendFindData(addr, hash, channel)

	for {
		n, _, _ := conn.ReadFromUDP(buf)
		container := &pb.Container{}
		proto.Unmarshal(buf[0:n], container)

		if container.REQUEST_TYPE != Request || container.REQUEST_ID != FindData ||
			container.GetRequestData().KEY != hash {
			t.Error("Didn't receive a FindContact request! Hmmm....")
			t.Fail()
		}
		return
	}
}
func TestNetwork_SendStoreData(t *testing.T) {
	nodeId := GenerateRandID(int64(rand.Intn(100)))
	rt := CreateAllWorkersForRoutingTable(K, IDLENGTH, 5, nodeId)
	network := InitializeNetwork(3, 12000, rt, true)

	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12000")
	conn, _ := net.ListenUDP("udp", addr)

	defer conn.Close()

	buf := make([]byte, 4096)

	channel := make(chan interface{})
	key := "00000"
	data := []byte("hello")
	go network.SendStoreData(addr, key, data, channel)

	for {
		n, _, _ := conn.ReadFromUDP(buf)
		container := &pb.Container{}
		proto.Unmarshal(buf[0:n], container)

		if container.REQUEST_TYPE != Request || container.REQUEST_ID != Store ||
			container.GetRequestStore().KEY != key || !(bytes.Equal(container.GetRequestStore().VALUE, data)) {
			t.Error("Didn't receive a FindContact request! Hmmm....")
			t.Fail()
		}
		return
	}
}

func TestNetwork_ReturnPing(t *testing.T) {
	msgID := "123"
	nodeId := GenerateRandID(int64(rand.Intn(100)))
	rt := CreateAllWorkersForRoutingTable(K, IDLENGTH, 5, nodeId)
	//addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12000")
	network := InitializeNetwork(5, 12000, rt, false)
	answerChannel := make(chan interface{})
	network.putInQueue(msgID, answerChannel)

	myID := "qwe"
	Info := &pb.RETURN_PING{ID: myID}
	Data := &pb.Container_ReturnPing{ReturnPing: Info}
	container := &pb.Container{REQUEST_TYPE: Return, REQUEST_ID: Ping, MSG_ID: msgID, Attachment: Data}

	go network.ReturnHandler(container)

	answer := <-network.takeFromQueue(msgID)
	switch answer := answer.(type) {
	case bool: //false
		fmt.Println(answer) //timedOut
		break
	default:
		if answer != myID {
			t.Error("I did not return the right id")
			t.Fail()
		}
		break
	}
}
func TestNetwork_ReturnContact(t *testing.T) {
	contactID := "00001"
	msgID := "123"

	rt := CreateAllWorkersForRoutingTable(20, 5, 5, "00000")
	//routingT := CreateAllWorkersForRoutingTable(20,160,5,GenerateRandID())
	rt.GiveOrder(OrderForRoutingTable{ADD, AddressTriple{"127.0.0.1", "9000", "00001"}, false})
	rt.GiveOrder(OrderForRoutingTable{ADD, AddressTriple{"127.0.0.1", "9000", "00010"}, false})
	rt.GiveOrder(OrderForRoutingTable{ADD, AddressTriple{"127.0.0.1", "9000", "00100"}, false})
	rt.GiveOrder(OrderForRoutingTable{ADD, AddressTriple{"127.0.0.1", "9000", "01000"}, false})

	network := InitializeNetwork(5, 12343, rt, false)
	answerChannel := make(chan interface{})
	network.putInQueue(msgID, answerChannel)

	//---------------------------------------------------------------------------------
	closestContacts := network.rt.FindKClosest(contactID) //mux on this?
	contactListReply := []*pb.RETURN_CONTACTS_CONTACT_INFO{}
	for i := range closestContacts {
		contactReply := &pb.RETURN_CONTACTS_CONTACT_INFO{IP: closestContacts[i].Triple.Ip, PORT: closestContacts[i].Triple.Port,
			ID: closestContacts[i].Triple.Id}
		contactListReply = append(contactListReply, contactReply)
	}
	Info := &pb.RETURN_CONTACTS{ContactInfo: contactListReply}
	Data := &pb.Container_ReturnContacts{ReturnContacts: Info}
	container := &pb.Container{REQUEST_TYPE: Return, REQUEST_ID: FindContact, MSG_ID: msgID, Attachment: Data}
	//---------------------------------------------------------------------------------

	go network.ReturnHandler(container)

	//---------------------------------------------------------------------------------
	listOfContacts := []AddressTriple{}
	for i := range container.GetReturnContacts().ContactInfo {
		listOfContacts = append(listOfContacts,
			AddressTriple{Ip: container.GetReturnContacts().ContactInfo[i].IP,
				Port: container.GetReturnContacts().ContactInfo[i].PORT,
				Id:   container.GetReturnContacts().ContactInfo[i].ID})
	}
	//---------------------------------------------------------------------------------
	answer := <-network.takeFromQueue(msgID)

	switch answer := answer.(type) {
	case bool: //false
		fmt.Println(answer) //timedOut
		break
	default:
		if !(reflect.DeepEqual(answer, listOfContacts)) {
			t.Error("I did not return the right id")
			t.Fail()
		}
		break
	}

}
func TestNetwork_ReturnData(t *testing.T) {
	msgID := "123"
	nodeId := GenerateRandID(int64(rand.Intn(100)))
	rt := CreateAllWorkersForRoutingTable(K, IDLENGTH, 5, nodeId)
	//addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12000")
	network := InitializeNetwork(5, 23423, rt, false)
	answerChannel := make(chan interface{})
	network.putInQueue(msgID, answerChannel)

	//Value := fileUtilsKademlia.ReadFileFromOS(DataID)
	Value := []byte("hello")
	Info := &pb.RETURN_DATA{VALUE: Value}
	Data := &pb.Container_ReturnData{ReturnData: Info}
	container := &pb.Container{REQUEST_TYPE: Return, REQUEST_ID: FindData, MSG_ID: msgID, Attachment: Data}

	go network.ReturnHandler(container)

	answer := <-network.takeFromQueue(msgID)
	switch answer := answer.(type) {
	case bool: //false
		fmt.Println(answer) //timedOut
		break
	default:
		if !(reflect.DeepEqual(answer, Value)) {
			t.Error("I did not return the right id")
			t.Fail()
		}
		break
	}

}
func TestNetwork_ReturnStore(t *testing.T) {
	msgID := "123"
	nodeId := GenerateRandID(int64(rand.Intn(100)))
	rt := CreateAllWorkersForRoutingTable(K, IDLENGTH, 5, nodeId)

	network := InitializeNetwork(5, 23223, rt, false)
	answerChannel := make(chan interface{})
	network.putInQueue(msgID, answerChannel)

	Value := "Stored"
	Info := &pb.RETURN_STORE{VALUE: Value}
	Data := &pb.Container_ReturnStore{ReturnStore: Info}
	container := &pb.Container{REQUEST_TYPE: Return, REQUEST_ID: Store, MSG_ID: msgID, Attachment: Data}

	go network.ReturnHandler(container)

	answer := <-network.takeFromQueue(msgID)
	switch answer := answer.(type) {
	case bool: //false
		fmt.Println(answer) //timedOut
		break
	default:
		if answer != Value {
			t.Error("I did not return the right id")
			t.Fail()
		}
		break
	}

}

func TestPingReturned(t *testing.T) {

}
func TestContactReturned(t *testing.T) {

}
func TestDataReturned(t *testing.T) {

}
func TestStoreReturned(t *testing.T) {

}
