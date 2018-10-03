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
	myID := "qwerty"
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
			container.GetRequestPing().ID != myID {
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
			t.Error("Didn't receive a send data request! Hmmm....")
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
			t.Error("Didn't receive a store request! Hmmm....")
			t.Fail()
		}
		return
	}
}

func TestNetwork_ReturnPing(t *testing.T) {
	msgID := GenerateRandID(int64(rand.Intn(100)))
	nodeId := GenerateRandID(int64(rand.Intn(100)))
	rt := CreateAllWorkersForRoutingTable(K, IDLENGTH, 5, nodeId)

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
	msgID := GenerateRandID(int64(rand.Intn(100)))

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
			t.Error("I did not return the right list")
			t.Fail()
		}
		break
	}

}
func TestNetwork_ReturnData(t *testing.T) {
	msgID := GenerateRandID(int64(rand.Intn(100)))
	nodeId := GenerateRandID(int64(rand.Intn(100)))
	rt := CreateAllWorkersForRoutingTable(K, IDLENGTH, 5, nodeId)

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
			t.Error("I did not return the right data")
			t.Fail()
		}
		break
	}

}
func TestNetwork_ReturnStore(t *testing.T) {
	msgID := GenerateRandID(int64(rand.Intn(100)))
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
			t.Error("I did not return the right value")
			t.Fail()
		}
		break
	}

}

func TestPingReturned(t *testing.T) {
	myID := "qwerty"
	msgID := GenerateRandID(int64(rand.Intn(100)))
	nodeId := GenerateRandID(int64(rand.Intn(100)))

	rt := CreateAllWorkersForRoutingTable(K, IDLENGTH, 5, nodeId)
	network := InitializeNetwork(5, 12312, rt, false)

	Info := &pb.REQUEST_PING{ID: myID}
	Data := &pb.Container_RequestPing{RequestPing: Info}
	container := &pb.Container{REQUEST_TYPE: Request, REQUEST_ID: Ping, MSG_ID: msgID, Attachment: Data}

	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:12312")
	conn, _ := net.ListenUDP("udp", addr)
	defer conn.Close()
	buf := make([]byte, 4096)

	go network.RequestHandler(container, addr)

	for {
		n, _, _ := conn.ReadFromUDP(buf)
		container := &pb.Container{}
		proto.Unmarshal(buf[0:n], container)

		if container.REQUEST_TYPE != Return || container.REQUEST_ID != Ping {
			t.Error("Didn't receive a ping return! Hmmm....")
			t.Fail()
		}
		return
	}
}
func TestContactReturned(t *testing.T) {
	contactID := "00000"
	msgID := GenerateRandID(int64(rand.Intn(100)))
	nodeId := GenerateRandID(int64(rand.Intn(100)))

	rt := CreateAllWorkersForRoutingTable(K, IDLENGTH, 5, nodeId)
	network := InitializeNetwork(5, 31212, rt, false)

	Info := &pb.REQUEST_CONTACT{ID: contactID}
	Data := &pb.Container_RequestContact{RequestContact: Info}
	container := &pb.Container{REQUEST_TYPE: Request, REQUEST_ID: FindContact, MSG_ID: msgID, Attachment: Data}

	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:31212")
	conn, _ := net.ListenUDP("udp", addr)
	defer conn.Close()
	buf := make([]byte, 4096)

	go network.RequestHandler(container, addr)

	for {
		n, _, _ := conn.ReadFromUDP(buf)
		container := &pb.Container{}
		proto.Unmarshal(buf[0:n], container)

		if container.REQUEST_TYPE != Return || container.REQUEST_ID != FindContact {
			t.Error("Didn't receive a contact return! Hmmm....")
			t.Fail()
		}
		return
	}
}

/*  NEED to check for a real file in order to pass!!!!!!!
	Maybe two unit tests for this? one for file exist and one for return contacts!


func TestDataReturned(t *testing.T) {
	hash := GenerateRandID(int64(rand.Intn(100)))
	msgID := GenerateRandID(int64(rand.Intn(100)))
	nodeId := GenerateRandID(int64(rand.Intn(100)))

	rt := CreateAllWorkersForRoutingTable(K, IDLENGTH, 5, nodeId)
	network := InitializeNetwork(5, 31213, rt, false)

	Info := &pb.REQUEST_DATA{KEY: hash}   // <--- look in  a particular folder for a file with this "name"
	Data := &pb.Container_RequestData{RequestData: Info}
	container := &pb.Container{REQUEST_TYPE: Request, REQUEST_ID: FindData, MSG_ID: msgID, Attachment: Data}

	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:31213")
	conn, _ := net.ListenUDP("udp", addr)
	defer conn.Close()
	buf := make([]byte, 4096)

	go network.RequestHandler(container, addr)

	for {
		n, _, _ := conn.ReadFromUDP(buf)
		container := &pb.Container{}
		proto.Unmarshal(buf[0:n], container)

		if container.REQUEST_TYPE != Return || container.REQUEST_ID != FindData {
			t.Error("Didn't receive a data return! Hmmm....")
			t.Fail()
		}
		return
	}

}
 Same story here. Need to check for file on os.


func TestStoreReturned(t *testing.T) {
	KEY := GenerateRandID(int64(rand.Intn(100)))
	DATA := []byte("hello")
	msgID := GenerateRandID(int64(rand.Intn(100)))
	nodeId := GenerateRandID(int64(rand.Intn(100)))

	rt := CreateAllWorkersForRoutingTable(K, IDLENGTH, 5, nodeId)
	network := InitializeNetwork(5, 32213, rt, false)


	Info := &pb.REQUEST_STORE{KEY: KEY, VALUE: DATA}
	Data := &pb.Container_RequestStore{RequestStore: Info}
	container := &pb.Container{REQUEST_TYPE: Request, REQUEST_ID: Store, MSG_ID: msgID, Attachment: Data}

	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:32213")
	conn, _ := net.ListenUDP("udp", addr)
	defer conn.Close()
	buf := make([]byte, 4096)

	go network.RequestHandler(container, addr)

	for {
		n, _, _ := conn.ReadFromUDP(buf)
		container := &pb.Container{}
		proto.Unmarshal(buf[0:n], container)

		if container.REQUEST_TYPE != Return || container.REQUEST_ID != Store {
			t.Error("Didn't receive a store return! Hmmm....")
			t.Fail()
		}
		return
	}

}
*/
