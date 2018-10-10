package kdmlib

import (
	"Kademlia---P2P-DFS/github.com/golang/protobuf/proto"
	"Kademlia---P2P-DFS/kdmlib/fileutils"
	pb "Kademlia---P2P-DFS/kdmlib/proto_config"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
	"strings"
	"runtime"
	"os"
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
	kademlia    *Kademlia
	rt          RoutingTable
	serverConn  *net.UDPConn
	mux         *sync.Mutex
	queue       map[string]chan interface{} //<-change?
	timeLimit   int
	nodeID      string
}

type udpPacketAndInfo struct {
	address net.Addr
	n 		int
	packet 	[]byte
}



//launch the udp server on port "port", with the specified amount of workers. needs the routing table and file channel
func UdpServer(port int,numberOfWorkers int,ownId string,table RoutingTable,fileChannel chan fileUtilsKademlia.Order){
	pc, err := net.ListenPacket("udp", ":"+strconv.Itoa(port))
	packetsChan := make(chan udpPacketAndInfo,500)
	if err != nil {
		log.Fatal(err)
	}

	for i:=0;i<numberOfWorkers ;i++  {
		go ConnectionWorker(packetsChan,pc,ownId,table,fileChannel)
	}

	for {
		buff := make([]byte,2048)
		n,addr , _:= pc.ReadFrom(buff)
		packetsChan <- udpPacketAndInfo{n:n,address:addr,packet:buff}
	}
}


//reads from the channel and handles the packet
func ConnectionWorker(packetsChan chan udpPacketAndInfo, conn net.PacketConn , nodeId string , table RoutingTable , fileChannel chan fileUtilsKademlia.Order){
	for toto := range packetsChan {
		container := &pb.Container{}
		proto.Unmarshal(toto.packet[:toto.n],container)
		requestHandler(container,conn,toto.address,table,nodeId,fileChannel)
	}
}


//handle the container according to its id, and update routing table
func requestHandler(container *pb.Container,conn net.PacketConn,addr net.Addr , table RoutingTable,nodeId string,fileChannel chan fileUtilsKademlia.Order){
	switch container.ID {
	case Ping:
		conn.WriteTo([]byte("pong"),addr)
		break
	case FindContact:
		packet , err := proto.Marshal(handleFindContact(table,container.GetRequestContact().ID,nodeId))
		if err == nil{
			conn.WriteTo(packet,addr)
		}else{
			fmt.Println("something went wrong")
		}
		break
	case FindData:
		packet , err := proto.Marshal(handleFindData(container.GetRequestData().KEY,nodeId,table))
		if err ==nil{
			conn.WriteTo(packet,addr)
		}else{
			fmt.Println("something went wrong")
		}
	case Store:
		handleStore(container.GetRequestStore().KEY,container.GetRequestStore().VALUE,fileChannel)
		conn.WriteTo([]byte("stored"),addr)
	}
	table.GiveOrder(OrderForRoutingTable{Action:ADD,Target:AddressTriple{Ip:strings.Split(addr.String(),":")[0],Id:container.ID,Port:container.PORT},FromPinger:false})
}


//write the file to the file channel
func handleStore(name string, value []byte,fileChannel chan fileUtilsKademlia.Order){
	fileChannel <- fileUtilsKademlia.Order{Action:fileUtilsKademlia.ADD,Name:name,Content:value}
}

//check if data is present and returns it if it is. Returns a list of contacts if not present
func handleFindData(DataID string,nodeID string,table RoutingTable) *pb.Container{
	if fileUtilsKademlia.ReadFileFromOS(DataID) != nil {
		Value := fileUtilsKademlia.ReadFileFromOS(DataID)
		Info := &pb.RETURN_DATA{VALUE: Value}
		Data := &pb.Container_ReturnData{ReturnData: Info}
		Container := &pb.Container{REQUEST_TYPE: Return, REQUEST_ID: FindData, MSG_ID: "", ID: nodeID, Attachment: Data}
		return Container
	} else {
		return handleFindContact(table, DataID, nodeID)
	}
}

//returns list of contacts closest the contactId.
func handleFindContact(table RoutingTable,contactID string,nodeId string) *pb.Container{
	closestContacts := table.FindKClosest(contactID)
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
func sendPacket(ip string,port string,packet []byte) ([]byte , error){
	conn, err := net.Dial("udp", ip+":"+port)
	if err != nil {
		return nil , err
	}
	defer conn.Close()

	conn.Write(packet)
	conn.SetReadDeadline(time.Now().Add(time.Second*2))

	buff := make([]byte,2048)
	n,err := conn.Read(buff)
	if err==nil{
		return buff[:n],nil
	}else{
		return nil, err
	}
}

//send store request
func SendStore(distantIp string,distantId string,distantPort string,data []byte,dataName string,ownId string,ownPort string,table RoutingTable) (string , error ){
	fmt.Println("sending store request")
	msgID := GenerateRandID(int64(rand.Intn(100)))
	Info := &pb.REQUEST_STORE{KEY: dataName, VALUE: data}
	Data := &pb.Container_RequestStore{RequestStore: Info}
	Container := &pb.Container{REQUEST_TYPE: Request, REQUEST_ID: Store, MSG_ID: msgID, ID: ownId, Attachment: Data,PORT:ownPort}
	marshalled , _:= proto.Marshal(Container)
	answer,err := sendPacket(distantIp,distantPort,marshalled)
	if err !=nil{
		table.GiveOrder(OrderForRoutingTable{REMOVE,AddressTriple{Id:distantId,Ip:distantIp,Port:distantPort},false})
	}else{
		table.GiveOrder(OrderForRoutingTable{ADD,AddressTriple{Id:distantId,Ip:distantIp,Port:distantPort},false})
	}
	return string(answer),err
}

//send findnode request, return an address triple slice
func SendFindNode(distantIp string,distantId string,distantPort string,ownId string,ownPort string,table RoutingTable,idToLookFor string) ([]AddressTriple , error ){
	msgID := GenerateRandID(int64(rand.Intn(100)))
	Info := &pb.REQUEST_CONTACT{ID: idToLookFor}
	Data := &pb.Container_RequestContact{RequestContact: Info}
	Container := &pb.Container{REQUEST_TYPE: Request, REQUEST_ID: FindContact, MSG_ID: msgID, ID: ownId, PORT:ownPort,Attachment: Data}
	marshalled , _:= proto.Marshal(Container)
	answer,err := sendPacket(distantIp,distantPort,marshalled)

	if err !=nil{
		table.GiveOrder(OrderForRoutingTable{REMOVE,AddressTriple{Id:distantId,Ip:distantIp,Port:distantPort},false})
	}else{
		table.GiveOrder(OrderForRoutingTable{ADD,AddressTriple{Id:distantId,Ip:distantIp,Port:distantPort},false})
	}
	object := &pb.Container{}
	proto.Unmarshal(answer, object)
	result := make([]AddressTriple,len(object.GetReturnContacts().ContactInfo))
	for i:=0;i<len(object.GetReturnContacts().ContactInfo);i++{
		result[i]=AddressTriple{object.GetReturnContacts().ContactInfo[i].IP,object.GetReturnContacts().ContactInfo[i].PORT,object.GetReturnContacts().ContactInfo[i].ID}
	}

	return result,err
}

//send finddata request, if err = nil , first returned value is the data , if data == nil get address triple from the second value.
func SendFindData(distantIp string,distantId string,distantPort string,ownId string,ownPort string,table RoutingTable,idToLookFor string)([]byte, []AddressTriple , error){
	msgID := GenerateRandID(int64(rand.Intn(100)))
	Info := &pb.REQUEST_DATA{KEY: idToLookFor}
	Data := &pb.Container_RequestData{RequestData: Info}
	Container := &pb.Container{REQUEST_TYPE: Request, REQUEST_ID: FindData, MSG_ID: msgID, ID: ownId, Attachment: Data,PORT:ownPort}
	marshalled , _:= proto.Marshal(Container)
	answer,err := sendPacket(distantIp,distantPort,marshalled)
	if err !=nil{
		table.GiveOrder(OrderForRoutingTable{REMOVE,AddressTriple{Id:distantId,Ip:distantIp,Port:distantPort},false})
	}else{
		table.GiveOrder(OrderForRoutingTable{ADD,AddressTriple{Id:distantId,Ip:distantIp,Port:distantPort},false})
	}
	object := &pb.Container{}
	proto.Unmarshal(answer, object)
	if object.REQUEST_ID==FindContact{
		result := make([]AddressTriple,len(object.GetReturnContacts().ContactInfo))
		for i:=0;i<len(object.GetReturnContacts().ContactInfo);i++{
			result[i]=AddressTriple{object.GetReturnContacts().ContactInfo[i].IP,object.GetReturnContacts().ContactInfo[i].PORT,object.GetReturnContacts().ContactInfo[i].ID}
		}
		return answer,result,err
	}
	return answer, nil, err
}



// END OF JEREMY'S CODE
func InitializeNetwork(timeOutLimit int, port int, rt RoutingTable, nodeID string, test bool) *Network {
	network := &Network{}
	network.rt = rt
	network.mux = &sync.Mutex{}
	network.queue = make(map[string]chan interface{})
	network.timeLimit = timeOutLimit
	network.nodeID = nodeID

	if !test {
		network.UDPConnection(port)
	}
	return network
}
func (network *Network) UDPConnection(port int) {
	ServerAddr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(port))
	CheckError(err)

	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	network.serverConn = ServerConn
	buf := make([]byte, 1024)

	network.Listen(buf)

}
func (network *Network) Listen(buf []byte) { //pass ----> tasks []*workerPool.Task
	defer network.serverConn.Close()
	tasks := []*Task{}
	p := NewPool(tasks, runtime.NumCPU())
	for {
		//if len(tasks) >= 2 { //test
		go p.Run(network)
		//	tasks = []*Task{}
		//	}



		n, addr, err := network.serverConn.ReadFromUDP(buf)

		container := &pb.Container{}
		err = proto.Unmarshal(buf[0:n], container)
		if err != nil {
			fmt.Println("Error: ", err)
		}
		tasks = []*Task{}
		task := NewTask(func() error { return nil }, container, addr) //pass args such as container, addr
		tasks = append(tasks, task)
		p = NewPool(tasks, runtime.NumCPU())

		//go network.RequestHandler(container, addr) //remove this from here and call it from worker
	}
}
func (network *Network) RequestHandler(container *pb.Container, addr *net.UDPAddr) {
	switch container.REQUEST_ID {
	case Ping:
		network.ReturnPing(addr, container.MSG_ID)
		break
	case FindContact:
		network.ReturnContact(addr, container.MSG_ID, container.GetRequestContact().ID)
		break
	case FindData:
		network.ReturnData(addr, container.MSG_ID, container.GetRequestData().KEY) //need to pass a co
		break
	case Store:
		network.ReturnStore(addr, container.MSG_ID, container.GetRequestStore().KEY, container.GetRequestStore().VALUE)
		break
	default:
		fmt.Println("Something went horribly wrong! (Request)")
	}

	//network.rt.GiveOrder(OrderForRoutingTable{ADD,AddressTriple{addr.IP.String(), "9000", container.ID}, false})

}

//You ask something from someone!
func (network *Network) SendPing(addr *net.UDPAddr, returnChannel chan interface{}) {
	myID := network.nodeID
	msgID := GenerateRandID(int64(rand.Intn(100)))
	Info := &pb.REQUEST_PING{ID: myID}
	Data := &pb.Container_RequestPing{RequestPing: Info}
	Container := &pb.Container{REQUEST_TYPE: Request, REQUEST_ID: Ping, MSG_ID: msgID, ID: myID, Attachment: Data}
	network.putInQueue(msgID, returnChannel)
	network.RequestData(Container, addr)
}
func (network *Network) SendFindContact(addr *net.UDPAddr, contactID string, returnChannel chan interface{}) {
	msgID := GenerateRandID(int64(rand.Intn(100)))
	Info := &pb.REQUEST_CONTACT{ID: contactID}
	Data := &pb.Container_RequestContact{RequestContact: Info}
	Container := &pb.Container{REQUEST_TYPE: Request, REQUEST_ID: FindContact, MSG_ID: msgID, ID: network.nodeID, Attachment: Data}
	network.putInQueue(msgID, returnChannel)
	network.RequestData(Container, addr)
}
func (network *Network) SendFindData(addr *net.UDPAddr, hash string, returnChannel chan interface{}) {
	msgID := GenerateRandID(int64(rand.Intn(100)))
	Info := &pb.REQUEST_DATA{KEY: hash}
	Data := &pb.Container_RequestData{RequestData: Info}
	Container := &pb.Container{REQUEST_TYPE: Request, REQUEST_ID: FindData, MSG_ID: msgID, ID: network.nodeID, Attachment: Data}
	network.putInQueue(msgID, returnChannel)
	network.RequestData(Container, addr)
}
func (network *Network) SendStoreData(addr *net.UDPAddr, KEY string, DATA []byte, returnChannel chan interface{}) {
	msgID := GenerateRandID(int64(rand.Intn(100)))
	Info := &pb.REQUEST_STORE{KEY: KEY, VALUE: DATA}
	Data := &pb.Container_RequestStore{RequestStore: Info}
	Container := &pb.Container{REQUEST_TYPE: Request, REQUEST_ID: Store, MSG_ID: msgID, ID: network.nodeID, Attachment: Data}
	network.putInQueue(msgID, returnChannel)
	network.RequestData(Container, addr)
}

//Someone ask something from you and you return!
func (network *Network) ReturnPing(addr *net.UDPAddr, msgID string) {
	myID := network.nodeID
	Info := &pb.RETURN_PING{ID: myID}
	Data := &pb.Container_ReturnPing{ReturnPing: Info}
	Container := &pb.Container{REQUEST_TYPE: Return, REQUEST_ID: Ping, MSG_ID: msgID, ID: network.nodeID, Attachment: Data}
	network.ReturnRequestedData(Container, addr)
}
func (network *Network) ReturnContact(addr *net.UDPAddr, msgID string, contactID string) {
	closestContacts := network.rt.FindKClosest(contactID)
	contactListReply := []*pb.RETURN_CONTACTS_CONTACT_INFO{}
	for i := range closestContacts {
		contactReply := &pb.RETURN_CONTACTS_CONTACT_INFO{IP: closestContacts[i].Triple.Ip, PORT: closestContacts[i].Triple.Port,
			ID: closestContacts[i].Triple.Id}
		contactListReply = append(contactListReply, contactReply)
	}
	Info := &pb.RETURN_CONTACTS{ContactInfo: contactListReply}
	Data := &pb.Container_ReturnContacts{ReturnContacts: Info}
	Container := &pb.Container{REQUEST_TYPE: Return, REQUEST_ID: FindContact, MSG_ID: msgID, ID: network.nodeID, Attachment: Data}
	network.ReturnRequestedData(Container, addr)
}
func (network *Network) ReturnData(addr *net.UDPAddr, msgID string, DataID string) {

	if fileUtilsKademlia.ReadFileFromOS(DataID) != nil {
		Value := fileUtilsKademlia.ReadFileFromOS(DataID)
		Info := &pb.RETURN_DATA{VALUE: Value}
		Data := &pb.Container_ReturnData{ReturnData: Info}
		Container := &pb.Container{REQUEST_TYPE: Return, REQUEST_ID: FindData, MSG_ID: msgID, ID: network.nodeID, Attachment: Data}
		network.ReturnRequestedData(Container, addr)
	} else {
		network.ReturnContact(addr, msgID, network.kademlia.nodeId) //this should return my closest contacts? nodeID is my id?
	}
}
func (network *Network) ReturnStore(addr *net.UDPAddr, msgID string, key string, value []byte) { //work in progress

	network.fileChannel <- fileUtilsKademlia.Order{Action: fileUtilsKademlia.ADD, Name: key, Content: value}

	Info := &pb.RETURN_STORE{VALUE: "Stored"}
	Data := &pb.Container_ReturnStore{ReturnStore: Info}
	Container := &pb.Container{REQUEST_TYPE: Return, REQUEST_ID: Store, MSG_ID: msgID, ID: network.nodeID, Attachment: Data}
	network.ReturnRequestedData(Container, addr)
	//todo: implement tcp file transfer to replace this!
}

//Someone returns something you previously asked for!
func PingReturned(container *pb.Container, returnedRequest chan interface{}) {
	contactID := container.GetReturnPing().ID
	returnedRequest <- contactID
}
func ContactReturned(container *pb.Container, returnedRequest chan interface{}) {
	listOfContacts := []AddressTriple{}
	for i := range container.GetReturnContacts().ContactInfo {
		listOfContacts = append(listOfContacts,
			AddressTriple{Ip: container.GetReturnContacts().ContactInfo[i].IP,
				Port: container.GetReturnContacts().ContactInfo[i].PORT,
				Id:   container.GetReturnContacts().ContactInfo[i].ID})
	}

	returnedRequest <- listOfContacts
}
func DataReturned(container *pb.Container, returnedRequest chan interface{}) {
	Value := container.GetReturnData().VALUE
	returnedRequest <- Value
}
func StoreReturned(container *pb.Container, returnedRequest chan interface{}) {
	Value := container.GetReturnStore().VALUE
	returnedRequest <- Value
}

//helper functions
func (network *Network) RequestData(container *pb.Container, addr *net.UDPAddr) {

	conn, err := net.Dial("udp", addr.IP.String()+":"+strconv.Itoa(addr.Port))
	CheckError(err)
	conn.SetReadDeadline(time.Now().Add(time.Second * 3))
	buf := []byte(EncodeContainer(container))

	_, err = conn.Write(buf)
	buf = make([]byte, 1024)
	i, err := conn.Read(buf)

	container = &pb.Container{}
	err = proto.Unmarshal(buf[0:i], container)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	network.ReturnHandler(container)

}
func (network *Network) ReturnHandler(container *pb.Container) {
	returnedRequest := network.takeFromQueue(container.MSG_ID)
	if returnedRequest == nil {
		fmt.Println("Timeout")
		return
	}
	switch container.REQUEST_ID {
	case Ping:
		PingReturned(container, returnedRequest)
		break
	case FindContact:
		ContactReturned(container, returnedRequest)
		break
	case FindData:
		DataReturned(container, returnedRequest)
		break
	case Store:
		StoreReturned(container, returnedRequest)
		break
	default:
		fmt.Println("Something went horribly wrong! (Return)")
	}
	//network.rt.GiveOrder(OrderForRoutingTable{ADD, AddressTriple{addr.IP.String(), "9000", container.ID}, false})

}
func (network *Network) ReturnRequestedData(container *pb.Container, addr *net.UDPAddr) {

	buf := []byte(EncodeContainer(container))
	_, err := network.serverConn.WriteToUDP(buf, addr)

	if err != nil {
		fmt.Println(EncodeContainer(container), err)
	}
}
func (network *Network) putInQueue(msgID string, returnChannel chan interface{}) {
	network.mux.Lock()
	network.queue[msgID] = returnChannel
	network.mux.Unlock()
}
func (network *Network) takeFromQueue(msgID string) (returnedRequest chan interface{}) {
	network.mux.Lock()
	returnedRequest = network.queue[msgID]
	network.mux.Unlock()
	return returnedRequest
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

//---------------------------------------------TEST-------------------------------------------------------
type Task struct {
	Err       error
	container *pb.Container
	addr      *net.UDPAddr
	f         func() error
}

// NewTask initializes a new task based on a given work function.
func NewTask(f func() error, container *pb.Container, addr *net.UDPAddr) *Task {
	return &Task{f: f, container: container, addr: addr}
}

// Run runs a Task and does appropriate accounting via a given sync.WorkGroup.
func (t *Task) Run(wg *sync.WaitGroup) {
	t.Err = t.f()
	wg.Done()
}

// Pool is a worker group that runs a number of tasks at a configured concurrency.
type Pool struct {
	Tasks       []*Task
	concurrency int
	tasksChan   chan *Task
	wg          sync.WaitGroup
}

// NewPool initializes a new pool with the given tasks and at the given concurrency.
func NewPool(tasks []*Task, concurrency int) *Pool {
	return &Pool{
		Tasks:       tasks,
		concurrency: concurrency,
		tasksChan:   make(chan *Task),
	}
}

// HasErrors indicates whether there were any errors from tasks run. Its result
// is only meaningful after Run has been called.
func (p *Pool) HasErrors() bool {
	for _, task := range p.Tasks {
		if task.Err != nil {
			return true
		}
	}
	return false
}

// Run runs all work within the pool and blocks until it's finished.
func (p *Pool) Run(network *Network) {
	for i := 0; i < p.concurrency; i++ {
		go p.work(i, network)
	}

	p.wg.Add(len(p.Tasks))
	for _, task := range p.Tasks {
		p.tasksChan <- task
	}

	// all workers return
	close(p.tasksChan)
	p.wg.Wait()
}

// The work loop for any single goroutine.
func (p *Pool) work(i int, network *Network) { //this could be a requestHandler?
	for task := range p.tasksChan {
		fmt.Println("worker", i, "processing job", task.addr)
		//time.Sleep(time.Second*3)
		network.RequestHandler(task.container, task.addr)
		task.Run(&p.wg)
	}
}
