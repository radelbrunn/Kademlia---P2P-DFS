package kdmlib

import (
	"Kademlia---P2P-DFS/github.com/golang/protobuf/proto"
	"Kademlia---P2P-DFS/kdmlib/fileutils"
	pb "Kademlia---P2P-DFS/kdmlib/proto_config"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"
)

const (
	Request     = "Request"
	Return      = "Return"
	Ping        = "Ping"
	FindContact = "FindContact"
	FindData    = "FindData"
	Store       = "Store"
)

type Network struct {
	fileChannel   chan fileUtilsKademlia.Order
	pinnerChannel chan fileUtilsKademlia.Order
	packetsChan   chan udpPacketAndInfo
	rt            RoutingTable
	port          string
	serverConn    *net.UDPConn
	nodeID        string
	ip            string
	conn          net.PacketConn
}

func InitNetwork(port string, ip string, rt RoutingTable, nodeID string, test bool) *Network {
	network := &Network{}
	network.rt = rt
	network.port = port
	network.ip = ip
	network.nodeID = nodeID

	network.pinnerChannel, network.fileChannel, _ = fileUtilsKademlia.CreateAndLaunchFileWorkers()
	network.packetsChan = make(chan udpPacketAndInfo, 500)

	//Set test flag to true for testing puposes
	if !test {
		buffer := make([]byte, 4096)
		conn, err := net.ListenPacket("udp", network.ip+":"+network.port)
		network.conn = conn
		if err != nil {
			log.Fatal(err)
		}
		go network.UdpServer(ALPHA, buffer)
	}

	return network
}

type udpPacketAndInfo struct {
	address net.Addr
	n       int
	packet  []byte
}

//launch the udp server on port "port", with the specified amount of workers. needs the routing table and file channel
func (network *Network) UdpServer(numberOfWorkers int, buffer []byte) {

	for i := 0; i < numberOfWorkers; i++ {
		go network.ConnectionWorker()
	}

	defer network.conn.Close()

	for {
		n, addr, _ := network.conn.ReadFrom(buffer)
		network.packetsChan <- udpPacketAndInfo{n: n, address: addr, packet: buffer}
	}
}

//reads from the channel and handles the packet
func (network *Network) ConnectionWorker() {
	for toto := range network.packetsChan {
		if network.conn != nil {
			container := &pb.Container{}
			proto.Unmarshal(toto.packet[:toto.n], container)
			network.requestHandler(container, toto.address)
		}
	}
}

//handle the container according to its id, and update routing table
func (network *Network) requestHandler(container *pb.Container, addr net.Addr) {
	switch container.REQUEST_ID {
	case Ping:
		network.conn.WriteTo([]byte("pong"), addr)
		break
	case FindContact:
		packet, err := proto.Marshal(network.handleFindContact(container.GetRequestContact().ID))
		if err == nil {
			network.conn.WriteTo(packet, addr)
		} else {
			fmt.Println("something went wrong")
		}
		break
	case FindData:
		packet, err := proto.Marshal(network.handleFindData(container.GetRequestData().KEY))
		if err == nil {
			network.conn.WriteTo(packet, addr)
		} else {
			fmt.Println("something went wrong")
		}
	case Store:
		network.handleStore(network.nodeID+container.GetRequestStore().KEY, container.GetRequestStore().VALUE)
		fmt.Println("STORE SUCCEEDED: ", network.nodeID)
		network.conn.WriteTo([]byte("stored"), addr)
	}
	network.rt.GiveOrder(OrderForRoutingTable{Action: ADD, Target: AddressTriple{Ip: strings.Split(addr.String(), ":")[0], Id: container.ID, Port: container.PORT}, FromPinger: false})
}

//write the file to the file channel
func (network *Network) handleStore(name string, value []byte) {
	network.fileChannel <- fileUtilsKademlia.Order{Action: fileUtilsKademlia.ADD, Name: name, Content: value}
}

//check if data is present and returns it if it is. Returns a list of contacts if not present
func (network *Network) handleFindData(DataID string) *pb.Container {
	if fileUtilsKademlia.ReadFileFromOS(DataID) != nil {
		Value := fileUtilsKademlia.ReadFileFromOS(DataID)
		Info := &pb.RETURN_DATA{VALUE: Value}
		Data := &pb.Container_ReturnData{ReturnData: Info}
		Container := &pb.Container{REQUEST_TYPE: Return, REQUEST_ID: FindData, MSG_ID: "", ID: network.nodeID, Attachment: Data}
		return Container
	} else {
		return network.handleFindContact(DataID)
	}
}

//returns list of contacts closest the contactId.
func (network *Network) handleFindContact(contactID string) *pb.Container {
	closestContacts := network.rt.FindKClosest(contactID)
	contactListReply := []*pb.RETURN_CONTACTS_CONTACT_INFO{}
	for i := range closestContacts {
		contactReply := &pb.RETURN_CONTACTS_CONTACT_INFO{IP: closestContacts[i].Triple.Ip, PORT: closestContacts[i].Triple.Port,
			ID: closestContacts[i].Triple.Id}
		contactListReply = append(contactListReply, contactReply)
	}
	Info := &pb.RETURN_CONTACTS{ContactInfo: contactListReply}
	Data := &pb.Container_ReturnContacts{ReturnContacts: Info}
	Container := &pb.Container{REQUEST_TYPE: Return, REQUEST_ID: FindContact, MSG_ID: "", ID: network.nodeID, Attachment: Data}
	return Container
}

//sends a packet and returns the answers if any, returns error if time out
func sendPacket(ip string, port string, packet []byte) ([]byte, error) {
	conn, err := net.Dial("udp", ip+":"+port)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	conn.Write(packet)
	conn.SetReadDeadline(time.Now().Add(time.Second * 1))

	buff := make([]byte, 2048)
	n, err := conn.Read(buff)
	if err == nil {
		return buff[:n], nil
	} else {
		return nil, err
	}
}

//send store request
func (network *Network) SendStore(toContact AddressTriple, data []byte, fileName string) (string, error) {
	msgID := GenerateRandID(int64(rand.Intn(100)))
	Info := &pb.REQUEST_STORE{KEY: fileName, VALUE: data}
	Data := &pb.Container_RequestStore{RequestStore: Info}
	Container := &pb.Container{REQUEST_TYPE: Request, REQUEST_ID: Store, MSG_ID: msgID, ID: network.nodeID, Attachment: Data, PORT: network.port}
	marshaled, _ := proto.Marshal(Container)
	answer, err := sendPacket(toContact.Ip, toContact.Port, marshaled)
	if err != nil {
		network.rt.GiveOrder(OrderForRoutingTable{REMOVE, toContact, false})
	} else {
		network.rt.GiveOrder(OrderForRoutingTable{ADD, toContact, false})
	}
	return string(answer), err
}

//send findnode request, return an address triple slice
func (network *Network) SendFindNode(toContact AddressTriple, targetID string) ([]AddressTriple, error) {
	msgID := GenerateRandID(int64(rand.Intn(100)))
	Info := &pb.REQUEST_CONTACT{ID: targetID}
	Data := &pb.Container_RequestContact{RequestContact: Info}
	Container := &pb.Container{REQUEST_TYPE: Request, REQUEST_ID: FindContact, MSG_ID: msgID, ID: network.nodeID, PORT: network.port, Attachment: Data}
	marshaled, _ := proto.Marshal(Container)
	answer, err := sendPacket(toContact.Ip, toContact.Port, marshaled)
	if err != nil {
		network.rt.GiveOrder(OrderForRoutingTable{REMOVE, toContact, false})
		return nil, err
	} else {
		network.rt.GiveOrder(OrderForRoutingTable{ADD, toContact, false})
		object := &pb.Container{}
		proto.Unmarshal(answer, object)
		result := make([]AddressTriple, len(object.GetReturnContacts().ContactInfo))
		for i := 0; i < len(object.GetReturnContacts().ContactInfo); i++ {
			result[i] = AddressTriple{object.GetReturnContacts().ContactInfo[i].IP, object.GetReturnContacts().ContactInfo[i].PORT, object.GetReturnContacts().ContactInfo[i].ID}
		}
		return result, err
	}
}

//send finddata request, if err = nil , first returned value is the data , if data == nil get address triple from the second value.
func (network *Network) SendFindData(toContact AddressTriple, targetID string) ([]byte, []AddressTriple, error) {
	msgID := GenerateRandID(int64(rand.Intn(100)))
	Info := &pb.REQUEST_DATA{KEY: targetID}
	Data := &pb.Container_RequestData{RequestData: Info}
	Container := &pb.Container{REQUEST_TYPE: Request, REQUEST_ID: FindData, MSG_ID: msgID, ID: network.nodeID, Attachment: Data, PORT: network.port}
	marshaled, _ := proto.Marshal(Container)
	answer, err := sendPacket(toContact.Ip, toContact.Port, marshaled)
	if err != nil {
		network.rt.GiveOrder(OrderForRoutingTable{REMOVE, toContact, false})
	} else {
		network.rt.GiveOrder(OrderForRoutingTable{ADD, toContact, false})
	}
	object := &pb.Container{}
	proto.Unmarshal(answer, object)
	if object.REQUEST_ID == FindContact {
		result := make([]AddressTriple, len(object.GetReturnContacts().ContactInfo))
		for i := 0; i < len(object.GetReturnContacts().ContactInfo); i++ {
			result[i] = AddressTriple{object.GetReturnContacts().ContactInfo[i].IP, object.GetReturnContacts().ContactInfo[i].PORT, object.GetReturnContacts().ContactInfo[i].ID}
		}
		return answer, result, err
	}
	return answer, nil, err
}
