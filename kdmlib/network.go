package kdmlib

import (
	"Kademlia---P2P-DFS/github.com/golang/protobuf/proto"
	pb "Kademlia---P2P-DFS/kdmlib/proto_config"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
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
	kademlia *Kademlia
	rt       RoutingTable
}

func InitializeNetwork(port int, rt RoutingTable) *Network {
	network := &Network{}
	network.rt = rt
	network.UDPConnection(port)
	return network
}
func (network *Network) UDPConnection(port int) { //TODO: learn how to properly use channels
	ServerAddr, err := net.ResolveUDPAddr("udp", ":9000")
	CheckError(err)

	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)

	quit := make(chan struct{})
	go network.Listen(ServerConn, quit)
	<-quit
}
func (network *Network) Listen(ServerConn *net.UDPConn, quit chan struct{}) {
	//TODO: ADD worker pools
	buf := make([]byte, 1024)
	defer ServerConn.Close()
	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)
		container := &pb.Container{}
		err = proto.Unmarshal(buf[0:n], container)
		if err != nil {
			fmt.Println("Error: ", err)
		}
		network.Handler(container, addr)
	}
	quit <- struct{}{}
}
func (network *Network) Handler(container *pb.Container, addr *net.UDPAddr) {
	switch container.REQUEST_TYPE {
	case Request:
		switch container.REQUEST_ID {
		case Ping:
			fmt.Println("Received Ping_Request")
			//network.ReturnPing(addr)
			fmt.Println("Returned Ping_Request")
			break
		case FindContact:
			fmt.Println("Received FindContact_Request")
			fmt.Println(container.GetRequestContact().ID)
			//network.ReturnContactRequest(addr, container.GetRequestContact().ID)
			fmt.Println("Returned FindContact_Request")
			break
		case FindData:
			fmt.Println("Received FindData_Request")
			fmt.Println(container.GetRequestData().KEY)
			//network.ReturnDataRequest(addr, container.GetRequestData().KEY, container.GetRequestContact().ID)
			fmt.Println("Returned FindData_Request")
			break
		case Store:
			fmt.Println("Received Store_Request")
			fmt.Println(container.GetRequestStore().KEY)
			fmt.Println(container.GetRequestStore().VALUE)
			//network.ReturnStoreRequest(addr)
			fmt.Println("Returned Store_Request")
			break
		}
		break
	case Return:
		switch container.REQUEST_ID {
		case Ping:
			fmt.Println("Ping Returned")
			break
		case FindContact:
			fmt.Println("Contact Returned")
			break
		case FindData:
			fmt.Println("Data Returned")
			break
		case Store:
			fmt.Println("Store Returned")
			break
		}
		break
	}
}

func (network *Network) SendPing(addr *net.UDPAddr) {
	myID := network.kademlia.nodeId
	Info := &pb.REQUEST_PING{ID: myID}
	Data := &pb.Container_RequestPing{RequestPing: Info}
	Container := &pb.Container{REQUEST_TYPE: Request, REQUEST_ID: Ping, Attachment: Data}
	network.SendData(Container, addr)
}
func (network *Network) SendFindContactRequest(addr *net.UDPAddr, contactID string, returnChannel chan interface{}) {
	Info := &pb.REQUEST_CONTACT{ID: contactID}
	Data := &pb.Container_RequestContact{RequestContact: Info}
	Container := &pb.Container{REQUEST_TYPE: Request, REQUEST_ID: FindContact, Attachment: Data}
	network.SendData(Container, addr)
}
func (network *Network) SendFindDataRequest(addr *net.UDPAddr, hash string, returnChannel chan interface{}) {
	Info := &pb.REQUEST_DATA{KEY: hash}
	Data := &pb.Container_RequestData{RequestData: Info}
	Container := &pb.Container{REQUEST_TYPE: Request, REQUEST_ID: FindData, Attachment: Data}
	network.SendData(Container, addr)
}
func (network *Network) SendStoreRequest(addr *net.UDPAddr, KEY string, DATA []byte) {
	Info := &pb.REQUEST_STORE{KEY: KEY, VALUE: DATA}
	Data := &pb.Container_RequestStore{RequestStore: Info}
	Container := &pb.Container{REQUEST_TYPE: Request, REQUEST_ID: Store, Attachment: Data}
	network.SendData(Container, addr)
}
func (network *Network) ReturnPing(addr *net.UDPAddr) {
	myID := network.kademlia.nodeId
	Info := &pb.RETURN_PING{ID: myID}
	Data := &pb.Container_ReturnPing{ReturnPing: Info}
	Container := &pb.Container{REQUEST_TYPE: Return, REQUEST_ID: Ping, Attachment: Data}
	network.SendData(Container, addr)
}
func (network *Network) ReturnFindContactRequest(addr *net.UDPAddr, contactID string) {
	closestContacts := network.rt.FindKClosest(contactID)
	contactListReply := []*pb.RETURN_CONTACTS_CONTACT_INFO{}
	for i := range closestContacts {
		contactReply := &pb.RETURN_CONTACTS_CONTACT_INFO{IP: closestContacts[i].Triple.Ip, PORT: closestContacts[i].Triple.Port,
			ID: closestContacts[i].Triple.Id}
		contactListReply = append(contactListReply, contactReply)
	}
	Info := &pb.RETURN_CONTACTS{ContactInfo: contactListReply}
	Data := &pb.Container_ReturnContacts{ReturnContacts: Info}
	Container := &pb.Container{REQUEST_TYPE: Return, REQUEST_ID: FindContact, Attachment: Data}
	network.SendData(Container, addr)

}
func (network *Network) ReturnFindDataRequest(addr *net.UDPAddr, DataID string, contactID string) {

	//if network.node.fileUtils.ReadFileFromOS(DataID) !=nil {

	//Value:= network.node.fileUtils.getValue(DataID)
	Value := ""
	Info := &pb.RETURN_DATA{VALUE: Value}
	Data := &pb.Container_ReturnData{ReturnData: Info}
	Container := &pb.Container{REQUEST_TYPE: Return, REQUEST_ID: Ping, Attachment: Data}
	network.SendData(Container, addr)
	//} else{
	//	network.ReturnContactRequest(addr,contactID)
	//}
}
func (network *Network) ReturnStoreRequest(addr *net.UDPAddr) {
	//check if file already exist, if not, download and reply on the store request.
}
func (network *Network) SendData(container *pb.Container, contact *net.UDPAddr) {

	conn, err := net.Dial("udp", contact.IP.String()+":"+strconv.Itoa(contact.Port))
	CheckError(err)

	buf := []byte(EncodeContainer(container))
	_, err = conn.Write(buf)

	if err != nil {
		fmt.Println(EncodeContainer(container), err)
	}
	defer conn.Close()
}
func EncodeContainer(pack *pb.Container) []byte {
	data, err := proto.Marshal(pack)
	if err != nil {
		log.Fatal("marshalling error: ", err)
	}
	return data
}
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}

/*	testing of the router-table
	/*id:= "1010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010"
	contact := AddressTriple{"127.0.0.1", "9000", GenerateRandID()}
	contact2 := AddressTriple{"127.0.0.1", "9000", id}
	routingtable := CreateAllWorkersForRoutingTable(K, 160, 5, GenerateRandID())
	routingtable.GiveOrder(OrderForRoutingTable{ADD, contact, false})
	routingtable.GiveOrder(OrderForRoutingTable{ADD, contact2, false})

	contactss:=routingtable.FindKClosest(id)
	for _, element := range contactss{
		fmt.Println(element)
	}

*/
