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
	closest           []AddressTriple
	askedClosest      []AddressTriple
	gotResultBack     []AddressTriple
	fileChannel       chan fileUtilsKademlia.Order
	nodeId            string
	rt                RoutingTable
	network           Network
	alpha             int
	k                 int
	noCloserNodeCalls int
	exitThreshold     int
}

// Initializes a Kademlia struct
func NewKademliaInstance(nw *Network, nodeId string, alpha int, k int, rt RoutingTable) *Kademlia {
	kademlia := &Kademlia{}
	kademlia.network = *nw
	kademlia.nodeId = nodeId
	kademlia.rt = rt
	kademlia.alpha = alpha
	kademlia.k = k
	kademlia.noCloserNodeCalls = 0
	kademlia.exitThreshold = 3

	return kademlia
}

//A struct for sending Lookup orders
type LookupOrder struct {
	LookupType int
	Contact    AddressTriple
	Target     string
}

//Listener of the answerChannel
//Returns either a list of Contact or data
func (kademlia *Kademlia) AnswerListener(resultChannel chan interface{}) ([]AddressTriple, []byte) {
	for {
		select {
		case answer := <-resultChannel:
			switch answer := answer.(type) {
			case []AddressTriple:
				fmt.Println("Answer: ", answer)
				return answer, nil
			}
		}
	}
}

//Performs operations when the slice of contacts comes back from the network
func (kademlia *Kademlia) HandleContactAnswer(order LookupOrder, answerList []AddressTriple, resultChannel chan interface{}, lookupChannel chan LookupOrder) {
	if len(answerList) != 0 {
		//Refresh the list of closest contacts, according to the answer
		kademlia.RefreshClosest(answerList, order.Target)

		//If no closer node has been found in past "kademlia.exitThreshold" calls, write to the answerChannel (i.e. "return")
		//If not, ask next node from the list of closest
		if kademlia.noCloserNodeCalls > kademlia.exitThreshold {
			fmt.Println("Contacts found (no closer contact has been found in a while)")
			resultChannel <- kademlia.closest
		} else {
			kademlia.AskNextContact(order.Target, order.LookupType, lookupChannel)
		}
	} else {
		fmt.Println("No contacts returned")
		kademlia.noCloserNodeCalls++
		kademlia.AskNextContact(order.Target, order.LookupType, lookupChannel)
	}
}

//User by the Lookup function to perform FIND_NODE and FIND_DATA RPC calls
func (kademlia *Kademlia) LookupWorker(routineId int, lookupChannel chan LookupOrder, resultChannel chan interface{}) {
	fmt.Println("Goroutine ", routineId, " started...")

	//Execute orders from the channel
	for order := range lookupChannel {

		fmt.Println("Order: ", order)
		switch order.LookupType {

		case CONTACT_LOOKUP:
			//Send a FIND_NODE RPC to the contact
			contacts, err := kademlia.network.SendFindNode(order.Contact, order.Target)

			//Check if an error has occurred (typically the case on-timeout)
			if err == nil {
				//Handle the operations in a separate function
				kademlia.HandleContactAnswer(order, contacts, resultChannel, lookupChannel)
			} else {
				fmt.Println("TIMEOUT")
				kademlia.AskNextContact(order.Target, order.LookupType, lookupChannel)
			}

		case DATA_LOOKUP:
			//Send a FIND_DATA RPC to the contact
			data, contacts, err := kademlia.network.SendFindData(order.Contact, order.Target)

			if err == nil {
				if data != nil {
					//If some data is found,  write to the answerChannel (i.e. "return")
					resultChannel <- data
				} else {
					kademlia.HandleContactAnswer(order, contacts, resultChannel, lookupChannel)
				}
			} else {
				fmt.Println("TIMEOUT")
				kademlia.AskNextContact(order.Target, order.LookupType, lookupChannel)
			}
		}

		//Once the network has returned desired values, the node can be added to the list of nodes, which have responded/timed out
		kademlia.gotResultBack = append(kademlia.gotResultBack, order.Contact)

		//Check if all nodes have been asked and if all nodes have responded/timed out
		if kademlia.AskedAllContacts() && len(resultChannel) == 0 && len(kademlia.gotResultBack) == len(kademlia.askedClosest) {
			fmt.Println("Asked all len:", len(lookupChannel))
			resultChannel <- kademlia.closest
		}
	}
}

// Returns up to K closest contacts to the target contact.
// Uses worker pools for asking nodes
// Stops if same answer is received multiple times or if all contacts in kademlia.closest have been asked.
func (kademlia *Kademlia) LookupContact(target string, lookupType int) ([]AddressTriple, []byte) {

	//Instantiate channels for lookupWorkers and answers
	lookupChannel := make(chan LookupOrder, kademlia.alpha)
	resultChannel := make(chan interface{}, kademlia.k)

	//Instantiate lists of contacts
	kademlia.closest = []AddressTriple{}
	kademlia.askedClosest = []AddressTriple{}
	kademlia.gotResultBack = []AddressTriple{}

	//Append Triples from TripleAndDistance array to the slice of closest
	for _, e := range kademlia.rt.FindKClosest(target) {
		kademlia.closest = append(kademlia.closest, e.Triple)
	}

	fmt.Println(kademlia.closest)

	//Start at most Alpha Lookup goroutines
	for i := 0; i < kademlia.alpha && i < len(kademlia.closest); i++ {
		go kademlia.LookupWorker(i, lookupChannel, resultChannel)
	}

	//Loop through the closest contacts from the routing table and pass an order to the lookup channel
	for i := 0; i < kademlia.alpha && i < len(kademlia.closest); i++ {
		//Send an order to channel
		lookupChannel <- LookupOrder{lookupType, kademlia.closest[i], target}
		//Mark node as "asked" by appending it to the list of asked nodes
		kademlia.askedClosest = append(kademlia.askedClosest, kademlia.closest[i])
	}

	//Start a listener function, which returns the desired answer
	return kademlia.AnswerListener(resultChannel)

}

func (kademlia *Kademlia) LookupData(fileName string, test bool) (success bool) {
	fileNameHash := HashKademliaID(fileName)

	//Set test for tests with smaller IDs (for development purposes)
	if test {
		fileNameHash = fileName
	}

	_, data := kademlia.LookupContact(fileNameHash, DATA_LOOKUP)
	if data != nil {
		//TODO: implement file handling
		fmt.Println("File located")
		return true
	} else {
		fmt.Println("File could not be located")
		return false
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
// All nodes that doesn't already exist in kademlia.closest will be appended and then sorted
// If no new AddressTriple is added to kademlia.closest and no closer node has been found, "kademlia.noCloserNodeCalls" is incremented
func (kademlia *Kademlia) RefreshClosest(newContacts []AddressTriple, target string) {
	closestSoFar := kademlia.closest[0]
	elementsAlreadyPresent := true

	//Check for new contacts
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

	//Sort only if new elements have been appended
	if !elementsAlreadyPresent {
		kademlia.SortContacts(target)
	}

	//Check if any closer elements have been found
	if !elementsAlreadyPresent && kademlia.closest[0].Id != closestSoFar.Id {
		kademlia.noCloserNodeCalls = 0
	} else {
		kademlia.noCloserNodeCalls++
	}
}

//Sorts the list of closest contacts, according to distance to target, slices off the tail if more than K nodes are present
func (kademlia *Kademlia) SortContacts(target string) {
	sortedList := []AddressTriple{}

	//Go through elements one by one
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

	//Slice off the tail if more than K nodes are present
	if len(sortedList) > kademlia.k {
		sortedList = sortedList[:kademlia.k]
	}

	kademlia.closest = sortedList
}

//Checks if all contacts have been asked
func (kademlia *Kademlia) AskedAllContacts() (allAsked bool) {
	allAsked = true
	for i := range kademlia.closest {
		elementExists := false
		for j := range kademlia.askedClosest {
			if kademlia.closest[i].Id == kademlia.askedClosest[j].Id {
				elementExists = true
			}
		}
		if !elementExists {
			allAsked = false
		}
	}
	return allAsked
}

/*
func testRetContacts(toContact AddressTriple, targetID string) ([]AddressTriple, error) {
	time.Sleep(time.Second * 1)
	return []AddressTriple{toContact}, nil
}


//contacts, err := testRetContacts(order.Contact, order.Target)
*/
