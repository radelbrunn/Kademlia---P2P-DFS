package kdmlib

import (
	"Kademlia---P2P-DFS/github.com/golang/protobuf/proto"
	"Kademlia---P2P-DFS/kdmlib/fileutils"
	pb "Kademlia---P2P-DFS/kdmlib/proto_config"
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
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
	fileChannel    chan fileUtilsKademlia.Order
	pinnerChannel  chan fileUtilsKademlia.Order
	udpPacketsChan chan udpPacketAndInfo
	tcpPacketsChan chan tcpPacketAndInfo
	rt             RoutingTable
	port           string
	nodeID         string
	ip             string
	fileMap        fileUtilsKademlia.FileMap
	udpConn        net.PacketConn
	tcpConn        net.Listener
	fileTest       bool
}

//Initializes necessary variables and structs of the network, starts a UDP and a TCP server.
func InitNetwork(port string, ip string, rt RoutingTable, nodeID string, noNetworkTest bool, fileTest bool, fileChannel chan fileUtilsKademlia.Order, pinnerChannel chan fileUtilsKademlia.Order, fileMap fileUtilsKademlia.FileMap) *Network {
	network := &Network{}
	network.rt = rt
	network.port = port
	network.ip = ip
	network.nodeID = nodeID
	network.fileTest = fileTest

	network.fileChannel = fileChannel
	network.pinnerChannel = pinnerChannel
	network.fileMap = fileMap

	network.udpPacketsChan = make(chan udpPacketAndInfo, 500)
	network.tcpPacketsChan = make(chan tcpPacketAndInfo, 500)

	//Set test flag to true for testing puposes
	if !noNetworkTest {
		udpBuffer := make([]byte, 4096)
		tcpBuffer := make([]byte, 4096)

		//Initiate UDP connection (for handling the requests)
		udpConn, udpErr := net.ListenPacket("udp", network.ip+":"+network.port)
		if udpErr != nil {
			log.Fatal(udpErr)
		}
		network.udpConn = udpConn

		//Initiate TCP connection (for sending files)
		tcpConn, tcpErr := net.Listen("tcp", network.ip+":"+network.port)
		if tcpErr != nil {
			log.Fatal(tcpErr)
		}
		network.tcpConn = tcpConn

		go network.UDPServer(ALPHA, udpBuffer)
		go network.TCPServer(ALPHA, tcpBuffer)
	}

	return network
}

type udpPacketAndInfo struct {
	address net.Addr
	n       int
	packet  []byte
}

//Launch the UDP server on port "port", with the specified amount of workers.
func (network *Network) UDPServer(numberOfWorkers int, buffer []byte) {
	fmt.Println("UDP connection initiated for node '" + ConvertToHexAddr(network.nodeID) + "' on " + network.ip + ":" + network.port)

	for i := 0; i < numberOfWorkers; i++ {
		go network.UDPConnectionWorker()
	}

	defer network.udpConn.Close()

	for {
		n, addr, _ := network.udpConn.ReadFrom(buffer)
		network.udpPacketsChan <- udpPacketAndInfo{n: n, address: addr, packet: buffer}
	}
}

//Reads from the channel and handles the UDP packet.
func (network *Network) UDPConnectionWorker() {
	for toto := range network.udpPacketsChan {
		if network.udpConn != nil {
			container := &pb.Container{}
			proto.Unmarshal(toto.packet[:toto.n], container)
			network.requestHandler(container, toto.address)
		}
	}
}

type tcpPacketAndInfo struct {
	connection net.Conn
	packet     []byte
}

//Launch the TCP server on port "port", with the specified amount of workers.
func (network *Network) TCPServer(numberOfWorkers int, buffer []byte) {
	fmt.Println("TCP connection initiated for node '" + ConvertToHexAddr(network.nodeID) + "' on " + network.ip + ":" + network.port)

	for i := 0; i < numberOfWorkers; i++ {
		go network.TCPConnectionWorker()
	}

	defer network.tcpConn.Close()

	for {
		conn, err := network.tcpConn.Accept()
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
		network.tcpPacketsChan <- tcpPacketAndInfo{connection: conn, packet: buffer}
	}
}

//Reads from the channel and handles the TCP packet.
func (network *Network) TCPConnectionWorker() {
	for toto := range network.tcpPacketsChan {
		if network.tcpConn != nil {
			n, _ := toto.connection.Read(toto.packet)
			filename := string(toto.packet[:n])
			sendFileTCP(toto.connection, filename)
		}
	}
}

//Send a file via TCP (writes to a connection, which is provided as argument).
func sendFileTCP(conn net.Conn, fileName string) {
	defer conn.Close()

	file := fileUtilsKademlia.ReadFileFromOS(fileName)
	if file != nil {
		conn.Write(file)
	}
}

//Handle the container according to its ID and update routing table.
func (network *Network) requestHandler(container *pb.Container, addr net.Addr) {
	switch container.REQUEST_ID {
	case Ping:
		fmt.Println("Node '" + ConvertToHexAddr(network.nodeID) + "' received a PING RPC from node '" + ConvertToHexAddr(container.ID) + "'")
		network.udpConn.WriteTo([]byte("pong"), addr)
		break
	case FindContact:
		fmt.Println("Node '" + ConvertToHexAddr(network.nodeID) + "' received a FIND_NODE RPC from node '" + ConvertToHexAddr(container.ID) + "'")
		packet, err := proto.Marshal(network.handleFindContact(container.GetRequestContact().ID))
		if err == nil {
			network.udpConn.WriteTo(packet, addr)
		} else {
			fmt.Println("something went wrong")
		}
		break
	case FindData:
		fmt.Println("Node '" + ConvertToHexAddr(network.nodeID) + "' received a FIND_DATA RPC from node '" + ConvertToHexAddr(container.ID) + "'")
		packet, err := proto.Marshal(network.handleFindData(container.GetRequestData().KEY))
		if err == nil {
			network.udpConn.WriteTo(packet, addr)
		} else {
			fmt.Println("something went wrong")
		}
	case Store:
		fmt.Println("Node '" + ConvertToHexAddr(network.nodeID) + "' received a STORE RPC from node '" + ConvertToHexAddr(container.ID) + "'")
		network.handleStore(container.GetRequestStore().KEY, AddressTriple{addr.(*net.UDPAddr).IP.String(), container.PORT, container.ID})
		network.udpConn.WriteTo([]byte("stored"), addr)
	}
	network.rt.GiveOrder(OrderForRoutingTable{Action: ADD, Target: AddressTriple{Ip: strings.Split(addr.String(), ":")[0], Id: container.ID, Port: container.PORT}, FromPinger: false})
}

//Write the file to the file channel. File is fetched by TCP via RequestFile function.
func (network *Network) handleStore(fileName string, callbackContact AddressTriple) {
	data := network.RequestFile(callbackContact, fileName)
	if data != nil {
		//nodeID is appended to the fileName for testing if fileTest i set to true
		if network.fileTest {
			fileName = network.nodeID + fileName
		}
		network.fileChannel <- fileUtilsKademlia.Order{Action: fileUtilsKademlia.ADD, Name: fileName, Content: data}
	}
}

//Check if data is present and returns it if it is. Returns a list of contacts if not present.
func (network *Network) handleFindData(DataID string) *pb.Container {
	if network.fileMap.IsPresent(DataID) {
		Container := &pb.Container{REQUEST_TYPE: Return, REQUEST_ID: FindData, MSG_ID: "", ID: network.nodeID, Attachment: nil}
		return Container
	} else {
		return network.handleFindContact(DataID)
	}
}

//Returns list of contacts closest the contactID.
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

//Sends a packet and returns the answers if any, returns error if timeout.
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

//Send STORE request.
func (network *Network) SendStore(toContact AddressTriple, fileName string) (string, error) {
	fmt.Println("Node '" + ConvertToHexAddr(network.nodeID) + "' is sending a STORE RPC to node '" + ConvertToHexAddr(toContact.Id) + "'")
	msgID := GenerateRandID(int64(rand.Intn(100)), 160)
	Info := &pb.REQUEST_STORE{KEY: fileName, VALUE: nil}
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

//Send FIND_NODE request, return an AddressTriple slice.
func (network *Network) SendFindNode(toContact AddressTriple, targetID string) ([]AddressTriple, error) {
	fmt.Println("Node '" + ConvertToHexAddr(network.nodeID) + "' is sending a FIND_NODE RPC to node '" + ConvertToHexAddr(toContact.Id) + "'")
	msgID := GenerateRandID(int64(rand.Intn(100)), 160)
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

//Send FIND_DATA request.
//If data is found, it's details are returned in a AddressTriple, which can later be used to fetch file via TCP.
//If data is not located, a slice of AddressTriples is returned, containing closest contacts (equal to the result of a FIND_NODE RPC).
func (network *Network) SendFindData(toContact AddressTriple, targetID string) (AddressTriple, []AddressTriple, error) {
	fmt.Println("Node '" + ConvertToHexAddr(network.nodeID) + "' is sending a FIND_DATA RPC to node '" + ConvertToHexAddr(toContact.Id) + "'")
	msgID := GenerateRandID(int64(rand.Intn(100)), 160)
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
		fmt.Println("Contact: '" + ConvertToHexAddr(object.ID) + "' is sending back CONTACTS!")
		result := make([]AddressTriple, len(object.GetReturnContacts().ContactInfo))
		for i := 0; i < len(object.GetReturnContacts().ContactInfo); i++ {
			result[i] = AddressTriple{object.GetReturnContacts().ContactInfo[i].IP, object.GetReturnContacts().ContactInfo[i].PORT, object.GetReturnContacts().ContactInfo[i].ID}
		}
		return AddressTriple{}, result, err

	} else if object.REQUEST_ID == FindData {
		fmt.Println("Contact: '" + ConvertToHexAddr(object.ID) + "' is sending back its SIGNATURE!")
		return toContact, nil, err
	}

	return AddressTriple{}, nil, err
}

//Request a file via TCP.
func (network *Network) RequestFile(toContact AddressTriple, fileName string) []byte {
	fmt.Println("Node '" + ConvertToHexAddr(network.nodeID) + "' is requesting file " + fileName + " from node '" + ConvertToHexAddr(toContact.Id) + "' via TCP")
	conn, err := net.Dial("tcp", toContact.Ip+":"+toContact.Port)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	conn.Write([]byte(fileName))

	var buf bytes.Buffer
	io.Copy(&buf, conn)

	return buf.Bytes()
}
