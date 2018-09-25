package kdmlib

import (
	"container/list"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	ADD    = 1
	REMOVE = 2
	CACHE  = 3
)

//order to send to routing table
type OrderForRoutingTable struct {
	action     int
	target     AddressTriple
	fromPinger bool
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
type tripleAndDistance struct {
	Triple   AddressTriple
	Distance string
}

//returns a slice of the k closest node to id
func (routing routingTableAndCache) FindKClosest(id string) []tripleAndDistance {
	k := routing.k
	if len(id) != routing.idLength {
		fmt.Println("the id that you provided has a wrong size")
		return nil
	}
	nodes := make([]tripleAndDistance, len(id)*k)
	table := *routing.routingTable
	l := 0
	routing.lock.Lock()
	for i := 0; i < len(id); i++ {
		for j := table[i].Front(); j != nil; j = j.Next() {
			if table[i].Len() > 0 && j.Value != nil {
				triple := j.Value.(AddressTriple)
				distance, err := computeDistance(triple.Id, id)
				if err == nil {
					nodes[l] = tripleAndDistance{triple, distance}
					l++
				}
			}

		}
	}
	routing.lock.Unlock()
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Distance == "" {
			return false
		}
		return nodes[i].Distance < nodes[j].Distance
	})

	return nodes[:k]
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
		if order.action == ADD {
			index, _ = firstDifferentBit(order.target.Id, ownId)
		} else {
			index, _ = firstDifferentBit(order.target.Id, ownId)
		}
		//lock the table while modifying it
		routingTable.lock.Lock()
		if order.action == ADD {
			//if in table then same behavior than bump
			if isPresentInRoutingTable(routingTable, order.target, ownId) {
				for ele := (*(routingTable.routingTable))[index].Front(); ele != nil; ele = ele.Next() {
					if ele.Value.(AddressTriple).Id == order.target.Id {
						fmt.Println("already present so pushing this value")
						(*(routingTable.routingTable))[index].MoveToFront(ele)
						break
					}
				}
			} else {
				//if free space add it
				if (*(routingTable.routingTable))[index].Len() < k {
					(*(routingTable.routingTable))[index].PushFront(order.target)
				} else {
					//if order comes from a pinger dont send a new order to the pinger channel
					if !order.fromPinger {
						var last *list.Element
						for ele := (*(routingTable.routingTable))[index].Front(); ele != nil; ele = ele.Next() {
							last = ele
						}
						if len(pingerChannel) < cap(pingerChannel) {
							pingerChannel <- OrderForPinger{last.Value.(AddressTriple), order.target, true}
						} else {
							ordersToSend.PushBack(OrderForPinger{last.Value.(AddressTriple), order.target, true})
						}
					} else {
						(*(routingTable.cache))[index].PushFront(list.Element{})
					}
				}
			}
		} else if order.action == REMOVE {
			//(*(routingTable.routingTable))[index].Remove(order.target)
			for ele := (*(routingTable.routingTable))[index].Front(); ele != nil; ele = ele.Next() {
				if ele.Value.(AddressTriple).Id == order.target.Id {
					(*(routingTable.routingTable))[index].Remove(ele)
					break
				}
			}
		} else if order.action == CACHE {
			//(*(routingTable.cache))[index].PushFront(*(order.target))
		}

		routingTable.lock.Unlock()
	}

}

func ping(address AddressTriple) error {
	conn, err := net.Dial("udp", address.Ip+":"+address.Port)
	if err != nil {
		return err
	}
	defer conn.Close()

	//simple write
	conn.Write([]byte("ping"))

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
	i, _ := firstDifferentBit(ownid, triple.Id)
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
		return -1, errors.New("not the same length")
	}
	for i := 0; i < len(address1); i++ {
		letterfrom1 := address1[i]
		letterfrom2 := address2[i]
		if letterfrom1 != letterfrom2 {
			return i, nil
		}
	}
	return len(address1) - 1, nil
}

func computeDistance(id1 string, id2 string) (string, error) {
	if len(id1) != len(id2) {
		return "", errors.New("not the right distance")
	} else {
		var sb strings.Builder
		for i := 0; i < len(id1); i++ {
			if id1[i] == id2[i] {
				sb.WriteString("0")
			} else {
				sb.WriteString("1")
			}
		}
		return sb.String(), nil
	}
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

func (table RoutingTable) FindKClosest(id string) (tripleAndDistance []tripleAndDistance) {
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
