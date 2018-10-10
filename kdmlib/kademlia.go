package kdmlib

import (
	"Kademlia---P2P-DFS/kdmlib/fileutils"
	"fmt"
	"time"
)

const (
	DATA_LOOKUP    = 0
	CONTACT_LOOKUP = 1
)

type Kademlia struct {
	closest        []AddressTriple
	askedClosest   []AddressTriple
	gotResultBack  []AddressTriple
	fileChannel    chan fileUtilsKademlia.Order
	nodeId         string
	rt             RoutingTable
	network        Network
	alpha          int
	k              int
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
	kademlia.identicalCalls = 0
	kademlia.exitThreshold = 3

	return kademlia
}

func testRetContacts(toContact AddressTriple, targetID string) ([]AddressTriple, error) {
	time.Sleep(time.Second * 1)
	return []AddressTriple{toContact}, nil
}

type LookupOrder struct {
	LookupType int
	Contact    AddressTriple
	Target     string
}

func (kademlia *Kademlia) LookupWorker(routineId int, lookupChannel chan LookupOrder, resultChannel chan interface{}) {
	fmt.Println("Goroutine ", routineId, " started...")
	for order := range lookupChannel {
		if order.LookupType == CONTACT_LOOKUP {
			fmt.Println("Order: ", order)
			contacts, err := kademlia.network.SendFindNode(order.Contact, order.Target)
			//contacts, err := testRetContacts(order.Contact, order.Target)
			if err == nil {
				if len(contacts) != 0 {
					kademlia.RefreshClosest(contacts, order.Target)
					if kademlia.identicalCalls > kademlia.exitThreshold {
						fmt.Println("Contacts found (no closer contact has been found in a while)")
						resultChannel <- kademlia.closest
					} else {
						kademlia.AskNextContact(order.Target, order.LookupType, lookupChannel)
					}
				} else {
					fmt.Println("No contacts returned")
				}
			} else {
				fmt.Println("TIMEOUT")
				kademlia.AskNextContact(order.Target, order.LookupType, lookupChannel)
			}

			kademlia.gotResultBack = append(kademlia.gotResultBack, order.Contact)

			if kademlia.AskedAllContacts() && len(resultChannel) == 0 && len(kademlia.gotResultBack) == len(kademlia.askedClosest) {
				fmt.Println("Asked all len:", len(lookupChannel))
				resultChannel <- kademlia.closest
			}
		} else if order.LookupType == DATA_LOOKUP {
			fmt.Println("DATA_LOOKUP")
		}
	}
}

// Returns up to K closest contacts to the target contact.
// Uses worker pools for asking nodes
// Stops if same answer is received multiple times or if all contacts in kademlia.closest have been asked.
func (kademlia *Kademlia) LookupContact(target string, findData bool) ([]AddressTriple, string) {
	lookupChannel := make(chan LookupOrder, kademlia.alpha)
	resultChannel := make(chan interface{}, kademlia.alpha)

	kademlia.closest = []AddressTriple{}
	kademlia.askedClosest = []AddressTriple{}
	kademlia.gotResultBack = []AddressTriple{}

	//Append Triples from TripleAndDistance array to the slice of closest
	for _, e := range kademlia.rt.FindKClosest(target) {
		kademlia.closest = append(kademlia.closest, e.Triple)
	}

	fmt.Println(kademlia.closest)

	//Start at most Alpha lookup goroutines
	for i := 0; i < kademlia.alpha && i < len(kademlia.closest); i++ {
		go kademlia.LookupWorker(i, lookupChannel, resultChannel)
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
		case answer := <-resultChannel:
			switch answer := answer.(type) {

			//In case a slice of contacts is returned:
			//1. Update the list of closest contacts.
			//2. Check if there has not been any closer node for past "kademlia.exitThreshold" answers.
			//3. Continue by asking the next contact.
			case []AddressTriple:
				fmt.Println("Answer: ", answer)
				return answer, ""

				//In case a boolean value is returned (false)
				//Means the call to a contact has timed out:
				//Next contact is asked
			}

			//In case a string is returned:
			//Return
		}
	}
}

//Ask the next contact, which is fetched from kademlia.GetNextContact()
func (kademlia *Kademlia) AskNextContact(target string, lookupType int, lookupChannel chan LookupOrder) {
	nextContact := kademlia.GetNextContact()
	if nextContact != nil {
		fmt.Println("Next ", nextContact)
		lookupChannel <- LookupOrder{lookupType, *nextContact, target}
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

func (kademlia *Kademlia) AskedAllContacts() bool {
	contactsAlreadyPresent := true
	for i := range kademlia.closest {
		elementExists := false
		for j := range kademlia.askedClosest {
			if kademlia.closest[i].Id == kademlia.askedClosest[j].Id {
				elementExists = true
			}
		}
		if !elementExists {
			contactsAlreadyPresent = false
		}
	}
	return contactsAlreadyPresent
}
