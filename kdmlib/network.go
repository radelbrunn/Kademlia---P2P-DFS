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
	fileChannel chan fileUtilsKademlia.Order
	rt          RoutingTable
	port        string
	serverConn  *net.UDPConn
	nodeID      string
}

func InitializeNetwork(port string, rt RoutingTable, nodeID string, test bool) *Network {
	network := &Network{}
	network.rt = rt
	network.port = port
	network.nodeID = nodeID
	network.fileChannel = make(chan fileUtilsKademlia.Order)

	if !test {
		network.UdpServer(3, network.fileChannel)
	}

	return network
}

type udpPacketAndInfo struct {
	address net.Addr
	n       int
	packet  []byte
}

//launch the udp server on port "port", with the specified amount of workers. needs the routing table and file channel
func (network *Network) UdpServer(numberOfWorkers int, fileChannel chan fileUtilsKademlia.Order) {
	pc, err := net.ListenPacket("udp", ":"+network.port)
	packetsChan := make(chan udpPacketAndInfo, 500)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < numberOfWorkers; i++ {
		go network.ConnectionWorker(packetsChan, pc, network.nodeID, fileChannel)
	}

	for {
		buff := make([]byte, 2048)
		n, addr, _ := pc.ReadFrom(buff)
		packetsChan <- udpPacketAndInfo{n: n, address: addr, packet: buff}
	}
}

//reads from the channel and handles the packet
func (network *Network) ConnectionWorker(packetsChan chan udpPacketAndInfo, conn net.PacketConn, nodeId string, fileChannel chan fileUtilsKademlia.Order) {
	for toto := range packetsChan {
		container := &pb.Container{}
		proto.Unmarshal(toto.packet[:toto.n], container)
		network.requestHandler(container, conn, toto.address, nodeId, fileChannel)
	}
}

//handle the container according to its id, and update routing table
func (network *Network) requestHandler(container *pb.Container, conn net.PacketConn, addr net.Addr, nodeId string, fileChannel chan fileUtilsKademlia.Order) {
	switch container.REQUEST_ID {
	case Ping:
		conn.WriteTo([]byte("pong"), addr)
		break
	case FindContact:
		packet, err := proto.Marshal(network.handleFindContact(container.GetRequestContact().ID, nodeId))
		if err == nil {
			conn.WriteTo(packet, addr)
		} else {
			fmt.Println("something went wrong")
		}
		break
	case FindData:
		packet, err := proto.Marshal(network.handleFindData(container.GetRequestData().KEY, nodeId))
		if err == nil {
			conn.WriteTo(packet, addr)
		} else {
			fmt.Println("something went wrong")
		}
	case Store:
		network.handleStore(container.GetRequestStore().KEY, container.GetRequestStore().VALUE, fileChannel)
		conn.WriteTo([]byte("stored"), addr)
	}
	network.rt.GiveOrder(OrderForRoutingTable{Action: ADD, Target: AddressTriple{Ip: strings.Split(addr.String(), ":")[0], Id: container.ID, Port: container.PORT}, FromPinger: false})
}

//write the file to the file channel
func (network *Network) handleStore(name string, value []byte, fileChannel chan fileUtilsKademlia.Order) {
	fileChannel <- fileUtilsKademlia.Order{Action: fileUtilsKademlia.ADD, Name: name, Content: value}
}

//check if data is present and returns it if it is. Returns a list of contacts if not present
func (network *Network) handleFindData(DataID string, nodeID string) *pb.Container {
	if fileUtilsKademlia.ReadFileFromOS(DataID) != nil {
		Value := fileUtilsKademlia.ReadFileFromOS(DataID)
		Info := &pb.RETURN_DATA{VALUE: Value}
		Data := &pb.Container_ReturnData{ReturnData: Info}
		Container := &pb.Container{REQUEST_TYPE: Return, REQUEST_ID: FindData, MSG_ID: "", ID: nodeID, Attachment: Data}
		return Container
	} else {
		return network.handleFindContact(DataID, nodeID)
	}
}

//returns list of contacts closest the contactId.
func (network *Network) handleFindContact(contactID string, nodeId string) *pb.Container {
	closestContacts := network.rt.FindKClosest(contactID)
	contactListReply := []*pb.RETURN_CONTACTS_CONTACT_INFO{}
	for i := range closestContacts {
		contactReply := &pb.RETURN_CONTACTS_CONTACT_INFO{IP: closestContacts[i].Triple.Ip, PORT: closestContacts[i].Triple.Port,
			ID: closestContacts[i].Triple.Id}
		contactListReply = append(contactListReply, contactReply)
	}
	Info := &pb.RETURN_CONTACTS{ContactInfo: contactListReply}
	Data := &pb.Container_ReturnContacts{ReturnContacts: Info}
	Container := &pb.Container{REQUEST_TYPE: Return, REQUEST_ID: FindContact, MSG_ID: "", ID: nodeId, Attachment: Data}
	return Container
}

//sends a packet and returns the answers if any, returns error if time out
func (network *Network) sendPacket(ip string, port string, packet []byte) ([]byte, error) {
	conn, err := net.Dial("udp", ip+":"+port)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	conn.Write(packet)
	conn.SetReadDeadline(time.Now().Add(time.Second * 2))

	buff := make([]byte, 2048)
	n, err := conn.Read(buff)
	if err == nil {
		return buff[:n], nil
	} else {
		return nil, err
	}
}

//send store request
func (network *Network) SendStore(distantIp string, distantId string, distantPort string, data []byte, dataName string) (string, error) {
	fmt.Println("sending store request")
	msgID := GenerateRandID(int64(rand.Intn(100)))
	Info := &pb.REQUEST_STORE{KEY: dataName, VALUE: data}
	Data := &pb.Container_RequestStore{RequestStore: Info}
	Container := &pb.Container{REQUEST_TYPE: Request, REQUEST_ID: Store, MSG_ID: msgID, ID: network.nodeID, Attachment: Data, PORT: network.port}
	marshalled, _ := proto.Marshal(Container)
	answer, err := network.sendPacket(distantIp, distantPort, marshalled)
	if err != nil {
		network.rt.GiveOrder(OrderForRoutingTable{REMOVE, AddressTriple{Id: distantId, Ip: distantIp, Port: distantPort}, false})
	} else {
		network.rt.GiveOrder(OrderForRoutingTable{ADD, AddressTriple{Id: distantId, Ip: distantIp, Port: distantPort}, false})
	}
	return string(answer), err
}

//send findnode request, return an address triple slice
func (network *Network) SendFindNode(distantIp string, distantId string, distantPort string, targetID string) ([]AddressTriple, error) {
	msgID := GenerateRandID(int64(rand.Intn(100)))
	Info := &pb.REQUEST_CONTACT{ID: targetID}
	Data := &pb.Container_RequestContact{RequestContact: Info}
	Container := &pb.Container{REQUEST_TYPE: Request, REQUEST_ID: FindContact, MSG_ID: msgID, ID: network.nodeID, PORT: network.port, Attachment: Data}
	marshalled, _ := proto.Marshal(Container)
	answer, err := network.sendPacket(distantIp, distantPort, marshalled)

	if err != nil {
		network.rt.GiveOrder(OrderForRoutingTable{REMOVE, AddressTriple{Id: distantId, Ip: distantIp, Port: distantPort}, false})
	} else {
		network.rt.GiveOrder(OrderForRoutingTable{ADD, AddressTriple{Id: distantId, Ip: distantIp, Port: distantPort}, false})
	}
	object := &pb.Container{}
	proto.Unmarshal(answer, object)
	result := make([]AddressTriple, len(object.GetReturnContacts().ContactInfo))
	for i := 0; i < len(object.GetReturnContacts().ContactInfo); i++ {
		result[i] = AddressTriple{object.GetReturnContacts().ContactInfo[i].IP, object.GetReturnContacts().ContactInfo[i].PORT, object.GetReturnContacts().ContactInfo[i].ID}
	}

	return result, err
}

//send finddata request, if err = nil , first returned value is the data , if data == nil get address triple from the second value.
func (network *Network) SendFindData(distantIp string, distantId string, distantPort string, targetID string) ([]byte, []AddressTriple, error) {
	msgID := GenerateRandID(int64(rand.Intn(100)))
	Info := &pb.REQUEST_DATA{KEY: targetID}
	Data := &pb.Container_RequestData{RequestData: Info}
	Container := &pb.Container{REQUEST_TYPE: Request, REQUEST_ID: FindData, MSG_ID: msgID, ID: network.nodeID, Attachment: Data, PORT: network.port}
	marshalled, _ := proto.Marshal(Container)
	answer, err := network.sendPacket(distantIp, distantPort, marshalled)
	if err != nil {
		network.rt.GiveOrder(OrderForRoutingTable{REMOVE, AddressTriple{Id: distantId, Ip: distantIp, Port: distantPort}, false})
	} else {
		network.rt.GiveOrder(OrderForRoutingTable{ADD, AddressTriple{Id: distantId, Ip: distantIp, Port: distantPort}, false})
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
