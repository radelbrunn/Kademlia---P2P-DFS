package kdmlib

import (
	"Kademlia---P2P-DFS/kdmlib/fileutils"
	"fmt"
)

const (
	DATA_LOOKUP    = 0
	CONTACT_LOOKUP = 1
)

type Kademlia struct {
	closest            []AddressTriple
	askedClosest       []AddressTriple
	fileChannel        chan fileUtilsKademlia.Order
	nodeId             string
	rt                 RoutingTable
	network            Network
	alpha              int
	k                  int
	goroutines         int
	identicalCalls     int
	identicalThreshold int
}

// Initializes a Kademlia struct
func NewKademliaInstance(nw *Network, nodeId string, alpha int, k int, rt RoutingTable) *Kademlia {
	kademlia := &Kademlia{}
	kademlia.network = *nw
	kademlia.nodeId = nodeId
	kademlia.rt = rt
	kademlia.alpha = alpha
	kademlia.k = k
	kademlia.goroutines = 0
	kademlia.identicalCalls = 0
	kademlia.identicalThreshold = 10

	return kademlia
}

type LookupOrder struct {
	LookupType int
	Contact    AddressTriple
	Target     string
}

func (kademlia *Kademlia) LookupWorker(routineId int, lookupChannel <-chan LookupOrder, answerChannel chan interface{}) {
	fmt.Println("Goroutine ", routineId, " started...")
	for order := range lookupChannel {
		if order.LookupType == CONTACT_LOOKUP {
			fmt.Println("Lookup")
			answerChannel <- []AddressTriple{order.Contact}
			//kademlia.network.SendFindContact(ConvertToUDPAddr(order.Contact),order.Target, answerChannel)
		} else if order.LookupType == DATA_LOOKUP {
			//kademlia.network.SendFindData(ConvertToUDPAddr(order.Contact),order.Target, answerChannel)
		}
	}
}

// Returns up to K closest contacts to the target contact.
// Creates a channel and executes the calls to different nodes in separate goroutines
// Stops if same answer is received multiple times or if all contacts in kademlia.closest have been asked.
func (kademlia *Kademlia) LookupContact(target AddressTriple) []AddressTriple {
	lookupChannel := make(chan LookupOrder, kademlia.alpha)
	answerChannel := make(chan interface{}, kademlia.alpha)

	kademlia.closest = []AddressTriple{}
	kademlia.askedClosest = []AddressTriple{}

	//Append Triples from TripleAndDistance array to the slice of closest
	for _, e := range kademlia.rt.FindKClosest(target.Id) {
		kademlia.closest = append(kademlia.closest, e.Triple)
	}

	fmt.Println(kademlia.closest)

	//Start at most Alpha lookup goroutines
	for i := 0; i < kademlia.alpha && i < len(kademlia.closest); i++ {
		go kademlia.LookupWorker(i, lookupChannel, answerChannel)
	}

	//Loop through the closest contacts from the routing table
	for i := 0; i < kademlia.alpha && i < len(kademlia.closest); i++ {

		lookupChannel <- LookupOrder{CONTACT_LOOKUP, kademlia.closest[i], target.Id}

		kademlia.askedClosest = append(kademlia.askedClosest, kademlia.closest[i])
	}

	for {
		select {
		case answer := <-answerChannel:
			switch answer := answer.(type) {
			case []AddressTriple:
				fmt.Println(answer)
				kademlia.RefreshClosest(answer, target.Id)
				if kademlia.identicalCalls > kademlia.identicalThreshold {
					fmt.Println("Contacts found (multiple consecutive same answers)")
					return kademlia.closest
				} else {
					kademlia.AskNextContact(target.Id, lookupChannel, answerChannel)
				}
			}

		default:
			if kademlia.goroutines != 0 {
				return kademlia.closest
			}
		}
	}
}

//Ask the next contact, which is fetched from kademlia.GetNextContact()
func (kademlia *Kademlia) AskNextContact(target string, lookupChannel chan LookupOrder, answerChannel chan interface{}) {
	nextContact := kademlia.GetNextContact()
	if nextContact != nil {
		fmt.Println("Next ", nextContact)
		lookupChannel <- LookupOrder{CONTACT_LOOKUP, *nextContact, target}
		//go kademlia.network.SendFindContact(ConvertToUDPAddr(*nextNode), target, answerChannel)
	} else {
		fmt.Println("No more to ask")
	}
}

// Goes through the list of closest contacts and returns the next node to ask
func (kademlia *Kademlia) GetNextContact() *AddressTriple {
	for _, e := range kademlia.closest {
		if !AlreadyAsked(kademlia.askedClosest, e) {
			kademlia.askedClosest = append(kademlia.askedClosest, e)
			return &e
		}
	}
	return nil
}

// Refreshes the list of closest contacts
// All nodes that doesn't already exist in kademlia.closest will be appended and the sorted
// If no new AddressTriple is added to kademlia.closest, the kademlia.identicalCalls is incremented
func (kademlia *Kademlia) RefreshClosest(newContacts []AddressTriple, target string) {
	elementsAlreadyPresent := true
	for i := range newContacts {
		elementExists := false
		for j := range kademlia.closest {
			if kademlia.closest[j].Id == newContacts[i].Id {
				elementExists = true
			}
		}
		if !elementExists {
			elementsAlreadyPresent = false
			kademlia.closest = append(kademlia.closest, newContacts[i])
		}
	}

	if elementsAlreadyPresent {
		kademlia.identicalCalls++
	} else {
		kademlia.SortContacts(target)
		kademlia.identicalCalls = 0
	}

	//TODO: return only K closest ones (remove the tail)
	//kademlia.closest = kademlia.closest[:kademlia.k]
}

//Sorts the list of closest contacts, according to distance to target
func (kademlia *Kademlia) SortContacts(target string) {
	sortedList := []AddressTriple{}
	for i := range kademlia.closest {
		if len(sortedList) == 0 {
			sortedList = append(sortedList, kademlia.closest[i])
		} else {
			inserted := false
			for j := range sortedList {
				distA, _ := ComputeDistance(kademlia.closest[i].Id, target)
				distB, _ := ComputeDistance(sortedList[j].Id, target)
				if distA <= distB && !inserted {
					inserted = true
					sortedList = append(sortedList, AddressTriple{})
					copy(sortedList[j+1:], sortedList[j:])
					sortedList[j] = kademlia.closest[i]
				}
			}
			if !inserted {
				sortedList = append(sortedList, kademlia.closest[i])
			}
		}
	}
	kademlia.closest = sortedList
}
