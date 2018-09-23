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

//Work in progress. Much will change!
type Network struct {
	//should'nt node keep track of its own ID? unsure if it should be contained in kademlia!
	kademlia   *Kademlia
	node       *Node
	connection *net.UDPConn
}

const (
	Request     = "Request"
	Return      = "Return"
	Ping        = "Ping"
	FindContact = "FindContact"
	FindData    = "FindData"
	Store       = "Store"
)

func InitializeNetwork(port int) *Network {
	network := &Network{}
	network.UDPConnection(port)
	return network
}

func (network *Network) UDPConnection(port int) { //TODO: learn how to properly use channels
	ServerAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(port))
	CheckError(err)

	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	network.connection = ServerConn
	buf := make([]byte, 1024)

	quit := make(chan struct{})
	go network.Listen(buf, ServerAddr, quit)
	<-quit
}
func (network *Network) Listen(buf []byte, ServerAddr *net.UDPAddr, quit chan struct{}) {
	//TODO: ADD worker pools
	defer network.connection.Close()
	for {
		n, addr, err := network.connection.ReadFromUDP(buf)

		packet := &pb.Container{}
		err = proto.Unmarshal(buf[0:n], packet)
		if err != nil {
			fmt.Println("Error: ", err)
		}
		/*contact := AddressTriple{addr.IP.String(),addr.Port,packet.ID}
		  either use addr and ID to combine into an AddressTriple or just not use it!
		  tripleAndDistance does exactly what contact did? why change it? makes it so much harder to code!
		*/
		network.Handler(packet, addr)
	}
	quit <- struct{}{}
}
func (network *Network) Handler(container *pb.Container, addr *net.UDPAddr) {
	switch container.REQUEST_TYPE {
	case Request:
		switch container.REQUEST_ID {
		case Ping:
			network.ReturnPing(addr)
			break
		case FindContact:
			/*
				The recipient of the RPC returns up to k triples (IP address, port, nodeID)
				for the contacts that it knows to be closest to the key.
			*/
			network.ReturnContactRequest(addr)
			break
		case FindData:
			/*
				If a corresponding value is present on the recipient,the associated data is returned.
				Otherwise the RPC is equivalent to a FIND_NODE and a set of k triples is returned.
			*/
			network.ReturnDataRequest(addr)
			break
		case Store:
			network.ReturnStoreRequest(addr)
			break
		}
		break
	case Return:
		switch container.REQUEST_ID {
		case Ping:
			break
		case FindContact:
			break
		case FindData:
			break
		case Store:
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
	network.Send(Container, addr)
}
func (network *Network) SendContactRequest(addr *net.UDPAddr, contact_id string) {
	Info := &pb.REQUEST_CONTACT{ID: contact_id}
	Data := &pb.Container_RequestContact{RequestContact: Info}
	Container := &pb.Container{REQUEST_TYPE: Request, REQUEST_ID: FindContact, Attachment: Data}
	network.Send(Container, addr)
}
func (network *Network) SendDataRequest(addr *net.UDPAddr, HASH string) {
	Info := &pb.REQUEST_DATA{KEY: HASH}
	Data := &pb.Container_RequestData{RequestData: Info}
	Container := &pb.Container{REQUEST_TYPE: Request, REQUEST_ID: FindData, Attachment: Data}
	network.Send(Container, addr)
}
func (network *Network) SendStoreRequest(addr *net.UDPAddr, KEY string, DATA []byte) {
	Info := &pb.REQUEST_STORE{KEY: KEY, VALUE: DATA}
	Data := &pb.Container_RequestStore{RequestStore: Info}
	Container := &pb.Container{REQUEST_TYPE: Request, REQUEST_ID: Store, Attachment: Data}
	network.Send(Container, addr)
}

func (network *Network) ReturnPing(addr *net.UDPAddr) {
	myID := network.kademlia.nodeId
	Info := &pb.RETURN_PING{ID: myID}
	Data := &pb.Container_ReturnPing{ReturnPing: Info}
	Container := &pb.Container{REQUEST_TYPE: Return, REQUEST_ID: Ping, Attachment: Data}
	network.Send(Container, addr)
}
func (network *Network) ReturnContactRequest(addr *net.UDPAddr) {
	/*my_Contacts := network.node.rt.FindKClosest(contact.Id)
	Info := &pb.RETURN_CONTACTS{ContactInfo: my_Contacts}
	Data := &pb.Container_ReturnContacts{ReturnContacts: Info}
	Container := &pb.Container{REQUEST_TYPE: Return, REQUEST_ID: Ping, Attachment: Data}
	network.Send(Container, addr)
	*/
}
func (network *Network) ReturnDataRequest(addr *net.UDPAddr) {
	/*if(data){
	//data is available, return it
	}
	else{
	ReturnContactRequest(addr)
	}
	*/
}
func (network *Network) ReturnStoreRequest(addr *net.UDPAddr) {
	//check if file already exist, if not, download and reply on the store request.
}

func (network *Network) Send(container *pb.Container, addr *net.UDPAddr) {
	network.SendData(container, addr)
}
func (network *Network) SendData(container *pb.Container, addr *net.UDPAddr) {

	buf := []byte(EncodeContainer(container))
	_, err := network.connection.WriteToUDP(buf, addr)
	if err != nil {
		fmt.Println(EncodeContainer(container), err)
	}
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

func SendSomething(contact AddressTriple, conn *net.UDPConn) { //written for test purposes
	ServerAddr, err := net.ResolveUDPAddr("udp", contact.Ip+":"+contact.Port)
	CheckError(err)
	fmt.Println("beep")
	Info := &pb.REQUEST_PING{ID: "123"}
	Container := &pb.Container_RequestPing{RequestPing: Info}
	Data := &pb.Container{REQUEST_TYPE: Request, REQUEST_ID: Ping, Attachment: Container}
	Sendstuff(EncodeContainer(Data), conn, ServerAddr)

}
func Sendstuff(data []byte, conn *net.UDPConn, addr *net.UDPAddr) { //written for test purposes
	buf := []byte(data)
	_, err := conn.WriteToUDP(buf, addr)
	if err != nil {
		fmt.Println(data, err)
	}
}

//Kademlia Documentations
/*
	This RPC involves one node sending a PING message to another, which presumably replies with a PONG.
	This has a two-fold effect: the recipient of the PING must update the bucket corresponding to the sender;
	and, if there is a reply, the sender must update the bucket appropriate to the recipient.
	All RPC packets are required to carry an RPC identifier assigned by the sender and echoed in the reply.
	This is a quasi-random number of length B (160 bits).
*/ //Ping
/*
The FIND_NODE RPC includes a 160-bit key. The recipient of the RPC returns up to k triples (IP address, port, nodeID)
for the contacts that it knows to be closest to the key.
The recipient must return k triples if at all possible.
It may only return fewer than k if it is returning all of the contacts that it has knowledge of.
This is a primitive operation, not an iterative one.
*/ //Find_Contact
/*
A FIND_VALUE RPC includes a B=160-bit key. If a corresponding value is present on the recipient,
the associated data is returned. Otherwise the RPC is equivalent to a FIND_NODE and a set of k triples is returned.
This is a primitive operation, not an iterative one.
*/ //Find_data
/*
The sender of the STORE RPC provides a key and a block of data and requires that
the recipient store the data and make it available for later retrieval by that key (So just store it in rt?).
This is a primitive operation, not an iterative one.
*/ //Store
