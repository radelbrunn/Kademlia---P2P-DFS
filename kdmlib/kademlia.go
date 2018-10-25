package kdmlib

import (
	"Kademlia---P2P-DFS/kdmlib/fileutils"
	"fmt"
	"strconv"
	"sync"
)

const (
	DataLookup    = 0
	ContactLookup = 1
)

type Kademlia struct {
	closest           []AddressTriple
	askedClosest      []AddressTriple
	gotResultBack     []AddressTriple
	nodeID            string
	rt                RoutingTable
	network           Network
	alpha             int
	k                 int
	noCloserNodeCalls int
	exitThreshold     int
	lock              sync.Mutex
	fileMap           fileUtilsKademlia.FileMap
	fileChannel       chan fileUtilsKademlia.Order
}

//Initializes a Kademlia struct, along with its necessary variables
func NewKademliaInstance(nw *Network, nodeID string, alpha int, k int, rt RoutingTable, fileChannel chan fileUtilsKademlia.Order, fileMap fileUtilsKademlia.FileMap) *Kademlia {
	kademlia := &Kademlia{}
	kademlia.network = *nw
	kademlia.nodeID = nodeID
	kademlia.rt = rt
	kademlia.alpha = alpha
	kademlia.k = k
	kademlia.noCloserNodeCalls = 0
	kademlia.exitThreshold = 3
	kademlia.fileChannel = fileChannel
	kademlia.fileMap = fileMap

	return kademlia
}

//A struct for sending Lookup orders.
type LookupOrder struct {
	LookupType int
	Contact    AddressTriple
	Target     string
}

//Listener of the answerChannel.
//Returns either a slice of AddressTriples, or an AddressTriple, where the file is located (depending on the lookup type and outcomes).
func (kademlia *Kademlia) lookupListener(resultChannel chan interface{}) ([]AddressTriple, AddressTriple) {
	for {
		select {
		case answer := <-resultChannel:
			switch answer := answer.(type) {
			//Slice of AddressTriple is written to the channel in following scenarios:
			//1. Successful LookupContact
			//2. Successful LookupData, that was not able to locate data
			case []AddressTriple:
				PrintListOfContacts("CLOSEST CONTACTS TO NODE '"+ConvertToHexAddr(kademlia.nodeID)+"' ARE SUCCESSFULLY LOCATED!: ", answer)
				return answer, AddressTriple{}
				//A slice of bytes is only written to the channel in case the successful LookupData was able to found the file
			case AddressTriple:
				fmt.Println("CONTACT WITH DATA: '"+ConvertToHexAddr(answer.Id)+"'", answer.Ip+":"+answer.Port)
				return nil, answer
			}
		}
	}
}

//Performs operations when the slice of contacts comes back from the network.
func (kademlia *Kademlia) handleContactAnswer(order LookupOrder, answerList []AddressTriple, resultChannel chan interface{}, lookupWorkerChannel chan LookupOrder) {
	if len(answerList) != 0 {
		PrintListOfContacts("Got some contacts back from node '"+ConvertToHexAddr(order.Contact.Id)+"': ", answerList)
		//Refresh the list of closest contacts, according to the answer
		kademlia.refreshClosest(answerList, order.Target)

		//If no closer node has been found in past "kademlia.exitThreshold" calls, write to the answerChannel (i.e. "return")
		//If not, ask next node from the list of closest
		if kademlia.noCloserNodeCalls > kademlia.exitThreshold {
			fmt.Println("Contacts found (no closer contact has been found in a while)")
			if !IsResultClosed(resultChannel) {
				resultChannel <- kademlia.closest
			}
		} else {
			kademlia.askNextContact(order.Target, order.LookupType, lookupWorkerChannel)
		}
	} else {
		fmt.Println("No contacts returned")
		kademlia.noCloserNodeCalls++
		kademlia.askNextContact(order.Target, order.LookupType, lookupWorkerChannel)
	}
}

//User by the Lookup function to perform FIND_NODE and FIND_DATA RPC calls.
func (kademlia *Kademlia) lookupWorker(lookupWorkerChannel chan LookupOrder, resultChannel chan interface{}) {

	//Execute orders from the channel
	for order := range lookupWorkerChannel {

		switch order.LookupType {

		case ContactLookup:

			//Send a FIND_NODE RPC to the contact
			contacts, err := kademlia.network.SendFindNode(order.Contact, order.Target)

			//Check if an error has occurred (typically the case on-timeout)
			if err == nil {
				//Handle the operations in a separate function
				kademlia.handleContactAnswer(order, contacts, resultChannel, lookupWorkerChannel)
			} else {
				fmt.Println("FIND_NODE RPC from node '" + ConvertToHexAddr(kademlia.nodeID) + "' to node '" + ConvertToHexAddr(order.Contact.Id) + "' timed out")
				kademlia.askNextContact(order.Target, order.LookupType, lookupWorkerChannel)
			}

		case DataLookup:
			//Send a FIND_DATA RPC to the contact
			contactWithData, contacts, err := kademlia.network.SendFindData(order.Contact, order.Target)

			if err == nil {
				if contacts == nil {
					//If some data is found,  write to the answerChannel (i.e. "return")
					if !IsResultClosed(resultChannel) {
						resultChannel <- contactWithData
					}
				} else {
					kademlia.handleContactAnswer(order, contacts, resultChannel, lookupWorkerChannel)
				}
			} else {
				fmt.Println("FIND_DATA RPC from node '" + ConvertToHexAddr(kademlia.nodeID) + "' to node '" + ConvertToHexAddr(order.Contact.Id) + "' timed out")
				kademlia.askNextContact(order.Target, order.LookupType, lookupWorkerChannel)
			}
		}

		//Once the network has returned desired values, the node can be added to the list of nodes, which have responded/timed out
		kademlia.gotResultBack = append(kademlia.gotResultBack, order.Contact)

		//Check if all nodes have been asked and if all nodes have responded/timed out
		if kademlia.askedAllContacts() && len(resultChannel) == 0 {
			if !IsResultClosed(resultChannel) {
				resultChannel <- kademlia.closest
			}
		}
	}
}

//LookupAlgorithm initialization function
//Uses worker pools when sending queries to nodes
//Stops if same answer is received multiple times or if all contacts in "kademlia.closest" have been asked.
//Returns a slice of AddressTriples, or an AddressTriple, which has the desired file.
//Supports two lookupTypes: ContactLookup and DataLookup
func (kademlia *Kademlia) LookupAlgorithm(target string, lookupType int) ([]AddressTriple, AddressTriple) {

	fmt.Println("Node '" + ConvertToHexAddr(kademlia.nodeID) + "' has initiated Lookup Algorithm with parameter: " + strconv.Itoa(lookupType))

	//Instantiate channels for lookupWorkers and answers
	lookupWorkerChannel := make(chan LookupOrder, kademlia.alpha)
	resultChannel := make(chan interface{}, kademlia.k)

	//Instantiate lists of contacts
	kademlia.closest = []AddressTriple{}
	kademlia.askedClosest = []AddressTriple{}
	kademlia.gotResultBack = []AddressTriple{}

	//Append Triples from TripleAndDistance array to the slice of closest
	for _, e := range kademlia.rt.FindKClosest(target) {
		kademlia.closest = append(kademlia.closest, e.Triple)
	}

	PrintListOfContacts("ROUTING TABLE OF NODE '"+ConvertToHexAddr(kademlia.nodeID)+"': ", kademlia.closest)

	//Check if the list of closest is empty
	//If true, return nil
	if len(kademlia.closest) == 0 {
		fmt.Println("The routing table is empty. No contacts to ask.")
		return nil, AddressTriple{}
	}

	//Start at most Alpha lookup workers
	for i := 0; i < kademlia.alpha; i++ {
		go kademlia.lookupWorker(lookupWorkerChannel, resultChannel)
	}

	//Loop through the closest contacts from the routing table and pass an order to the lookup channel
	for i := 0; i < kademlia.alpha && i < len(kademlia.closest); i++ {
		//Send an order to channel
		lookupWorkerChannel <- LookupOrder{lookupType, kademlia.closest[i], target}
		//Mark node as "asked" by appending it to the list of asked nodes
		kademlia.askedClosest = append(kademlia.askedClosest, kademlia.closest[i])
	}

	for i := 0; i < kademlia.alpha-len(kademlia.closest); i++ {
		lookupWorkerChannel <- LookupOrder{lookupType, AddressTriple{"0", "00", GenerateZeroID(len(kademlia.nodeID))}, target}
	}

	//Start a listener function, which returns the desired answer
	contacts, contactWithData := kademlia.lookupListener(resultChannel)



	//Close the channels when the result is retrieved
	close(lookupWorkerChannel)
	close(resultChannel)

	switch lookupType {
	case ContactLookup:
		fmt.Println("FIND_NODE procedure, initiated by node '" + ConvertToHexAddr(kademlia.nodeID) + "' was finalized")
	case DataLookup:
		fmt.Println("FIND_DATA procedure, initiated by node '" + ConvertToHexAddr(kademlia.nodeID) + "' was finalized")
		if len(contactWithData.Id)!=0 {

			fmt.Println("contact which has file has id " + ConvertToHexAddr(contactWithData.Id))
		}else{
			fmt.Println("no contact with data found")
		}
	}

	return contacts, contactWithData

}

//Uses LookupAlgorithm to get the data with a fileName.
func (kademlia *Kademlia) LookupData(fileHash string) []byte {

	//Check if there is a contact, that has the file
	_, contactWithData := kademlia.LookupAlgorithm(fileHash, DataLookup)

	//If yes, download the file via TCP and store it locally.
	if contactWithData.Id != ""  {
		data := kademlia.network.RequestFile(contactWithData, fileHash)
		if data != nil {
			kademlia.fileChannel <- fileUtilsKademlia.Order{Action: fileUtilsKademlia.ADD, Name: fileHash, Content: data}
			fmt.Println("File located and downloaded")
			return data
		} else {
			fmt.Println("File could not be located")
			return nil
		}
	} else {
		fmt.Println("File could not be located")
		return nil
	}
}

//A struct for sending Store orders.
type StoreOrder struct {
	Contact  AddressTriple
	FileName string
}

//Used to send STORE RPC to nodes.
func (kademlia *Kademlia) storeWorker(storeWorkerChannel chan StoreOrder, resultChannel chan bool) {

	//Execute orders from the channel
	for order := range storeWorkerChannel {
		answer, err := kademlia.network.SendStore(order.Contact, order.FileName)

		if err == nil && answer == "stored" {
			fmt.Println("STORE RPC from node '" + ConvertToHexAddr(kademlia.nodeID) + "' to node '" + ConvertToHexAddr(order.Contact.Id) + "' SUCCEEDED")
			resultChannel <- true
		} else {
			fmt.Println("STORE RPC from node '" + ConvertToHexAddr(kademlia.nodeID) + "' to node '" + ConvertToHexAddr(order.Contact.Id) + "' FAILED")
			resultChannel <- false
		}
	}
}

//Listener of the storeResultChannel.
func (kademlia *Kademlia) storeListener(resultChannel chan bool, expectedNumAnswers int) {
	answersReturned := 0
	for {
		select {
		case <-resultChannel:
			answersReturned++
			//Return when all answers are received
			if answersReturned == expectedNumAnswers {
				fmt.Println("STORE procedure, initiated by node '" + ConvertToHexAddr(kademlia.nodeID) + "' was finalized")
				return
			}
		}
	}
}

//Finds K closest contacts and stores the file.
//Uses LookupContact to find closest contacts to hash of fileName.
func (kademlia *Kademlia) StoreData(fileName string) {
	fileNameHash := fileName

	//Check whether the file exists.
	//If yes, get the list of closest and send the file to these nodes.
	fmt.Println("entered store data")
	fmt.Println("file is present? ", kademlia.network.fileMap.IsPresent(fileName))
	fmt.Println("file id is "+fileName)
	if kademlia.network.fileMap.IsPresent(fileName) {
		contacts, _ := kademlia.LookupAlgorithm(fileNameHash, ContactLookup)
		if contacts != nil {

			//Instantiate channels for lookupWorkers and answers
			storeWorkerChannel := make(chan StoreOrder, kademlia.k)
			resultChannel := make(chan bool, kademlia.k)

			//Start at most Alpha store workers
			for i := 0; i < kademlia.alpha && i < len(contacts); i++ {
				go kademlia.storeWorker(storeWorkerChannel, resultChannel)
			}

			//Loop through the list of closest and send orders to the store channel
			for _, contact := range contacts {
				storeWorkerChannel <- StoreOrder{contact, fileName}
			}

			kademlia.storeListener(resultChannel, len(contacts))

		} else {
			fmt.Println("Contacts are empty! Something went wrong...")
		}

	} else {
		fmt.Println("File does not exist locally (nothing to store)")
	}

}

//Ask the next contact, which is fetched from kademlia.GetNextContact().
func (kademlia *Kademlia) askNextContact(target string, lookupType int, lookupWorkerChannel chan LookupOrder) {
	kademlia.lock.Lock()

	nextContact := kademlia.getNextContact()
	if nextContact != nil {
		kademlia.askedClosest = append(kademlia.askedClosest, *nextContact)
		if !IsLookupClosed(lookupWorkerChannel) {
			lookupWorkerChannel <- LookupOrder{lookupType, *nextContact, target}
		}
	}

	kademlia.lock.Unlock()
}

// Goes through the list of closest contacts and returns the next node to ask.
func (kademlia *Kademlia) getNextContact() *AddressTriple {
	for _, e := range kademlia.closest {
		if !AlreadyAsked(kademlia.askedClosest, e) {
			return &e
		}
	}
	return nil
}

// Refreshes the list of closest contacts
// All nodes that doesn't already exist in kademlia.closest will be appended and then sorted.
// If no new AddressTriple is added to kademlia.closest and no closer node has been found, "kademlia.noCloserNodeCalls" is incremented.
func (kademlia *Kademlia) refreshClosest(newContacts []AddressTriple, target string) {
	kademlia.lock.Lock()

	closestSoFar := kademlia.closest[0]
	elementsAlreadyPresent := true

	//Check for new contacts
	for i := range newContacts {
		elementExists := false
		if newContacts[i].Id != kademlia.nodeID {
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
	}

	//Sort only if new elements have been appended
	if !elementsAlreadyPresent {
		kademlia.sortContacts(target)
	}

	//Check if any closer elements have been found
	if !elementsAlreadyPresent && kademlia.closest[0].Id != closestSoFar.Id {
		kademlia.noCloserNodeCalls = 0
	} else {
		kademlia.noCloserNodeCalls++
	}

	kademlia.lock.Unlock()
}

//Sorts the list of closest contacts, according to distance to target, slices off the tail if more than K nodes are present.
func (kademlia *Kademlia) sortContacts(target string) {
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

//Checks if all contacts have been asked.
func (kademlia *Kademlia) askedAllContacts() (allAsked bool) {
	allAsked = true
	for i := range kademlia.closest {
		elementAsked := false
		elementResponded := false
		for j := range kademlia.askedClosest {
			if kademlia.closest[i].Id == kademlia.askedClosest[j].Id {
				elementAsked = true
			}
		}
		for j := range kademlia.gotResultBack {
			if kademlia.closest[i].Id == kademlia.gotResultBack[j].Id {
				elementResponded = true
			}
		}
		if !elementAsked || !elementResponded {
			allAsked = false
		}
	}
	return allAsked
}
