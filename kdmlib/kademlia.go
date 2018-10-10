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
	closest        []AddressTriple
	askedClosest   []AddressTriple
	fileChannel    chan fileUtilsKademlia.Order
	nodeId         string
	rt             RoutingTable
	network        Network
	alpha          int
	k              int
	goroutines     int
	identicalCalls int
	exitThreshold  int
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
	kademlia.exitThreshold = 3

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
			fmt.Println("CONTACT_LOOKUP")
			answerChannel <- []AddressTriple{order.Contact}
			//kademlia.network.SendFindContact(ConvertToUDPAddr(order.Contact),order.Target, answerChannel)
		} else if order.LookupType == DATA_LOOKUP {
			fmt.Println("DATA_LOOKUP")
			//kademlia.network.SendFindData(ConvertToUDPAddr(order.Contact),order.Target, answerChannel)
		}
	}
}

// Returns up to K closest contacts to the target contact.
// Uses worker pools for asking nodes
// Stops if same answer is received multiple times or if all contacts in kademlia.closest have been asked.
func (kademlia *Kademlia) LookupContact(target string, findData bool) ([]AddressTriple, string) {
	lookupChannel := make(chan LookupOrder, kademlia.alpha)
	answerChannel := make(chan interface{}, kademlia.alpha)

	kademlia.closest = []AddressTriple{}
	kademlia.askedClosest = []AddressTriple{}

	//Append Triples from TripleAndDistance array to the slice of closest
	for _, e := range kademlia.rt.FindKClosest(target) {
		kademlia.closest = append(kademlia.closest, e.Triple)
	}

	fmt.Println(kademlia.closest)

	//Start at most Alpha lookup goroutines
	for i := 0; i < kademlia.alpha && i < len(kademlia.closest); i++ {
		go kademlia.LookupWorker(i, lookupChannel, answerChannel)
	}

	//Loop through the closest contacts from the routing table
	for i := 0; i < kademlia.alpha && i < len(kademlia.closest); i++ {

		if !findData {
			lookupChannel <- LookupOrder{CONTACT_LOOKUP, kademlia.closest[i], target}
		} else {
			lookupChannel <- LookupOrder{DATA_LOOKUP, kademlia.closest[i], target}
		}

		kademlia.askedClosest = append(kademlia.askedClosest, kademlia.closest[i])
	}

	for {
		select {
		case answer := <-answerChannel:
			switch answer := answer.(type) {

			//In case a slice of contacts is returned:
			//1. Update the list of closest contacts.
			//2. Check if there has not been any closer node for past "kademlia.exitThreshold" answers.
			//3. Continue by asking the next contact.
			case []AddressTriple:
				kademlia.RefreshClosest(answer, target)
				if kademlia.identicalCalls > kademlia.exitThreshold {
					fmt.Println("Contacts found (no closer contact has been found in a while)")
					return kademlia.closest, ""
				} else {
					kademlia.AskNextContact(target, findData, lookupChannel, answerChannel)
				}

			//In case a boolean value is returned (false)
			//Means the call to a contact has timed out:
			//Next contact is asked
			case bool:
				kademlia.AskNextContact(target, findData, lookupChannel, answerChannel)
			}

			//In case a string is returned:
			//Return
		}
	}
}

//Ask the next contact, which is fetched from kademlia.GetNextContact()
func (kademlia *Kademlia) AskNextContact(target string, findData bool, lookupChannel chan LookupOrder, answerChannel chan interface{}) {
	nextContact := kademlia.GetNextContact()
	if nextContact != nil {
		fmt.Println("Next ", nextContact)
		if !findData {
			lookupChannel <- LookupOrder{CONTACT_LOOKUP, *nextContact, target}
		} else {
			lookupChannel <- LookupOrder{DATA_LOOKUP, *nextContact, target}

		}
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

	if len(kademlia.closest) > kademlia.k {
		kademlia.closest = kademlia.closest[:kademlia.k]
	}

	//Check if I have smth closer than b4
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
