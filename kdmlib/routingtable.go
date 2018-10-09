package kdmlib

import (
	"container/list"
	"errors"
	"fmt"
	"net"
	"sort"
	pb "Kademlia---P2P-DFS/kdmlib/proto_config"
	"sync"
	"time"
	"math/rand"
)

const (
	ADD    = 1
	REMOVE = 2
	CACHE  = 3
)

//order to send to routing table
type OrderForRoutingTable struct {
	Action     int
	Target     AddressTriple
	FromPinger bool
}

//address and id of a node
type AddressTriple struct {
	Ip   string
	Port string
	Id   string
}

//order to send to a ping worker
type OrderForPinger struct {
	toPing      AddressTriple
	newElement  AddressTriple
	shouldCache bool
}

//routing table , cache and muutex
type routingTableAndCache struct {
	routingTable *[]list.List
	cache        *[]list.List
	lock         sync.Mutex
	k            int
	idLength     int
}

//triple and distance
type TripleAndDistance struct {
	Triple   AddressTriple
	Distance string
}

//returns a slice of the k closest node to id
func (routing routingTableAndCache) FindKClosest(id string) []TripleAndDistance {
	k := routing.k
	if len(id) != routing.idLength {
		fmt.Println("the id that you provided has a wrong size")
		return nil
	}
	nodes := make([]TripleAndDistance, len(id)*k)
	table := *routing.routingTable
	l := 0
	counter := 0
	routing.lock.Lock()
	for i := 0; i < len(id); i++ {
		for j := table[i].Front(); j != nil; j = j.Next() {
			counter++
			if table[i].Len() > 0 && j.Value != nil {
				triple := j.Value.(AddressTriple)
				distance, err := ComputeDistance(triple.Id, id)
				if err == nil && len(distance) > 0 {
					nodes[l] = TripleAndDistance{triple, distance}
					l++
				}
			}

		}
	}
	routing.lock.Unlock()
	sort.Slice(nodes, func(i, j int) bool {
		if len(nodes[i].Distance) == 0 {
			return false
		} else if len(nodes[j].Distance) == 0 {
			return true
		}
		return nodes[i].Distance < nodes[j].Distance
	})

	if l < k {
		return nodes[:l]
	} else {
		return nodes[:k]
	}
}

//read from in, ping address from this channel and send appropriate order to routing table
func pingWorker(in chan OrderForPinger, out chan OrderForRoutingTable, chanLocker *sync.Mutex) {
	for order := range in {
		leastRecentlySeen := order.toPing
		err := ping(leastRecentlySeen)
		//lock the channel to ensure that the second order arrives right after the first
		chanLocker.Lock()

		//if no answer from ping remove dead node and add new node
		if err != nil {
			fmt.Println("there was an error")
			out <- OrderForRoutingTable{REMOVE, leastRecentlySeen, true}
			out <- OrderForRoutingTable{ADD, order.newElement, true}
		} else {
			//bump the node who responded
			fmt.Println("we got a response yay")
			out <- OrderForRoutingTable{ADD, leastRecentlySeen, true}
			if order.shouldCache {
				//if there is a new element , cache in cache table
				out <- OrderForRoutingTable{CACHE, order.newElement, true}
			}
		}
		chanLocker.Unlock()
	}
}

//returns the least recently seen node for the corresponding kbucket
//func getLeastRecentlySeen(routingTable routingTableAndCache, subtreeIndex int) *list.Element {
//	//lock the table to ensure that there is no problem while reading it
//	routingTable.lock.Lock()
//	res := (*(routingTable.routingTable))[subtreeIndex].Back()
//	routingTable.lock.Unlock()
//	return res
//}

//read from the channel and updates the routing table accordingly
func updateRoutingTableWorker(routingTable routingTableAndCache, channel chan OrderForRoutingTable, ownId string, k int, pingerChannel chan OrderForPinger) {
	ordersToSend := list.New()
	for order := range channel {
		//fill the pinger channel if it is not full
		for ordersToSend.Len() > 0 && (cap(pingerChannel) > len(pingerChannel)) {
			pingerChannel <- ordersToSend.Remove(ordersToSend.Front()).(OrderForPinger)
		}
		//var element list.Element
		//element = *(order.target)
		//fmt.Println(element.Value.(string))
		var index int

		index, errorLength := firstDifferentBit(order.Target.Id, ownId)
		//discard if id is not ok
		if errorLength == nil {
			//lock the table while modifying it
			routingTable.lock.Lock()
			if order.Action == ADD {
				//if in table then same behavior than bump
				if isPresentInRoutingTable(routingTable, order.Target, ownId) {
					bumpElement(routingTable, index, order)
				} else {
					//if free space add it
					if (*(routingTable.routingTable))[index].Len() < k {
						(*(routingTable.routingTable))[index].PushFront(order.Target)
					} else {
						//if order comes from a pinger dont send a new order to the pinger channel
						if !order.FromPinger {
							sendToPinger(routingTable, index, order, pingerChannel, ordersToSend)
						}
					}
				}
			} else if order.Action == REMOVE {
				removeFromRoutingTable(routingTable, index, order)
			} else if order.Action == CACHE {
				//(*(routingTable.cache))[index].PushFront(*(order.target))
			}

			routingTable.lock.Unlock()
		} else {
			fmt.Println(errorLength)
		}

	}

}

func sendToPinger(routingTable routingTableAndCache, index int, order OrderForRoutingTable, pingerChannel chan OrderForPinger, ordersToSend *list.List) {
	var last *list.Element
	for ele := (*(routingTable.routingTable))[index].Front(); ele != nil; ele = ele.Next() {
		if ele.Value != nil {
			last = ele
		}
	}
	if len(pingerChannel) < cap(pingerChannel) {
		pingerChannel <- OrderForPinger{last.Value.(AddressTriple), order.Target, true}
	} else {
		ordersToSend.PushBack(OrderForPinger{last.Value.(AddressTriple), order.Target, true})
	}
}

func bumpElement(routingTable routingTableAndCache, index int, order OrderForRoutingTable) {
	for ele := (*(routingTable.routingTable))[index].Front(); ele != nil; ele = ele.Next() {
		if ele.Value.(AddressTriple).Id == order.Target.Id {
			fmt.Println("already present so pushing this value")
			(*(routingTable.routingTable))[index].MoveToFront(ele)
			break
		}
	}
}

func removeFromRoutingTable(routingTable routingTableAndCache, index int, order OrderForRoutingTable) {
	for ele := (*(routingTable.routingTable))[index].Front(); ele != nil; ele = ele.Next() {
		if ele.Value.(AddressTriple).Id == order.Target.Id {
			(*(routingTable.routingTable))[index].Remove(ele)
			break
		}
	}

}

func ping(address AddressTriple) error {
	conn, err := net.Dial("udp", address.Ip+":"+address.Port)
	if err != nil {
		return err
	}
	defer conn.Close()

	msgID := GenerateRandID(int64(rand.Intn(100)))
	Info := &pb.REQUEST_PING{ID: "asdasd"}
	Data := &pb.Container_RequestPing{RequestPing: Info}
	Container := &pb.Container{REQUEST_TYPE: Request, REQUEST_ID: Ping, MSG_ID: msgID, Attachment: Data}

	//simple write
	conn.Write([]byte(EncodeContainer(Container)))

	conn.SetReadDeadline(time.Now().Add(time.Second * 2))
	//simple Read
	buffer := make([]byte, 1024)
	answer, err := conn.Read(buffer)
	fmt.Println(string(buffer[:answer]))
	//return an non nil error if the node doesn't pong back
	//return nil if the node answers
	return err
}

//create a new routing table
func createRoutingTable(k int, idLength int) routingTableAndCache {

	a := make([]list.List, idLength)
	b := make([]list.List, idLength)
	for i := 0; i < idLength; i++ {
		a[i] = *list.New()
		b[i] = *list.New()
	}
	return routingTableAndCache{&a, &b, sync.Mutex{}, k, idLength}
}

func isPresentInRoutingTable(routingTable routingTableAndCache, triple AddressTriple, ownid string) bool {
	i, err := firstDifferentBit(ownid, triple.Id)
	if err != nil {
		fmt.Println(err)
		return false
	}
	for j := (*(routingTable.routingTable))[i].Front(); j != nil; j = j.Next() {
		value, ok := j.Value.(AddressTriple)
		id := value.Id
		if ok && id == triple.Id {
			(*(routingTable.routingTable))[i].MoveToFront(j)
			return true
		}
	}
	return false
}

//return the position of the first different bit
func firstDifferentBit(address1 string, address2 string) (int, error) {
	if len(address1) != len(address2) {
		return -1, errors.New("lengths of the IDs are different")
	}
	for i := 0; i < len(address1); i++ {
		letterFrom1 := address1[i]
		letterFrom2 := address2[i]
		if letterFrom1 != letterFrom2 {
			return i, nil
		}
	}
	return len(address1) - 1, nil
}

type RoutingTable struct {
	routingChannel chan OrderForRoutingTable
	lock           *sync.Mutex
	routingtable   routingTableAndCache
}

func (table RoutingTable) GiveOrder(order OrderForRoutingTable) {
	table.lock.Lock()
	table.routingChannel <- order
	table.lock.Unlock()
}

func (table RoutingTable) FindKClosest(id string) (tripleAndDistance []TripleAndDistance) {
	return table.routingtable.FindKClosest(id)
}

func CreateAllWorkersForRoutingTable(k int, idLegnth int, numberOfPinger int, ownId string) RoutingTable {
	routingChannel := make(chan OrderForRoutingTable, 1000)
	pingingChannel := make(chan OrderForPinger, 1000)
	routingTable := createRoutingTable(k, idLegnth)
	var channelLocker = &sync.Mutex{}

	go updateRoutingTableWorker(routingTable, routingChannel, ownId, k, pingingChannel)
	for i := 0; i < numberOfPinger; i++ {
		go pingWorker(pingingChannel, routingChannel, channelLocker)
	}

	return RoutingTable{routingChannel, channelLocker, routingTable}
}
