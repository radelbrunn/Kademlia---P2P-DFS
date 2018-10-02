package kdmlib

import (
	"Kademlia---P2P-DFS/kdmlib/fileutils"
	"fmt"
)

type Kademlia struct {
	closest            []AddressTriple
	fileChannel        chan fileUtilsKademlia.Order
	asked              map[string]bool
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
	kademlia.asked = make(map[string]bool)
	kademlia.nodeId = nodeId
	kademlia.rt = rt
	kademlia.alpha = alpha
	kademlia.k = k
	kademlia.goroutines = 0
	kademlia.identicalCalls = 0
	kademlia.identicalThreshold = 2

	return kademlia
}

//TODO add timeout handling (bool channel input??)
// Returns up to K closest contacts to the target contact.
// Creates a channel and executes the calls to different nodes in separate goroutines
// Stops if same answer is received multiple times or if all contacts in kademlia.closest have been asked.
func (kademlia *Kademlia) LookupContact(target *AddressTriple) []AddressTriple {
	answerChannel := make(chan interface{}, kademlia.alpha)
	kademlia.closest = []AddressTriple{}

	for _, e := range kademlia.rt.FindKClosest(target.Id) {
		kademlia.closest = append(kademlia.closest, e.Triple)
	}

	//Loop through the closest contacts from the routing table
	for i := 0; i < kademlia.alpha && i < len(kademlia.closest); i++ {
		fmt.Println("Sending find contact message to node")
		kademlia.goroutines++
		go kademlia.network.SendFindContact(ConvertToUDPAddr(kademlia.closest[i]), target.Id, answerChannel)

		kademlia.asked[kademlia.closest[i].Id] = true
	}

	//Channel listener
	for {
		select {
		case answer := <-answerChannel:
			switch answer := answer.(type) {
			case []AddressTriple:
				kademlia.RefreshClosest(answer, target.Id)
				if kademlia.identicalCalls > kademlia.identicalThreshold {
					fmt.Println("Contacts found (multiple consecutive same answers)")
					return kademlia.closest
				} else {
					nextNode := kademlia.GetNextNode()
					if nextNode != nil {
						fmt.Println("Sending find contact message to node")
						go kademlia.network.SendFindContact(ConvertToUDPAddr(*nextNode), target.Id, answerChannel)
					} else {
						fmt.Println("Thread ended")
						kademlia.goroutines--
					}
				}
			}
		default:
			if kademlia.goroutines == 0 {
				return kademlia.closest
			}
			if kademlia.goroutines < kademlia.k {
				nextNode := kademlia.GetNextNode()
				if nextNode != nil {
					fmt.Println("Sending find contact message to node")
					kademlia.goroutines++
					go kademlia.network.SendFindContact(ConvertToUDPAddr(*nextNode), target.Id, answerChannel)
				}
			}
		}
	}
}

//TODO add timeout handling (bool channel input??)
// Returns the file, according to the hashed filename, or a list of closest contacts, if a file is not found.
// Creates a channel and executes the calls to different nodes in separate goroutines
// Stops if same answer is received multiple times (contacts), if all contacts in kademlia.closest have been asked or if the data is located.
func (kademlia *Kademlia) LookupData(hash string) ([]AddressTriple, string) {
	answerChannel := make(chan interface{}, kademlia.alpha)
	kademlia.closest = []AddressTriple{}

	for _, e := range kademlia.rt.FindKClosest(hash) {
		kademlia.closest = append(kademlia.closest, e.Triple)
	}

	//Loop through the closest contacts from the routing table
	for i := 0; i < kademlia.alpha && i < len(kademlia.closest); i++ {
		fmt.Println("Sending find data message")
		kademlia.goroutines++
		go kademlia.network.SendFindData(ConvertToUDPAddr(kademlia.closest[i]), hash, answerChannel)

		kademlia.asked[kademlia.closest[i].Id] = true
	}

	//Channel listener
	for {
		select {
		case answer := <-answerChannel:
			switch answer := answer.(type) {
			case string:
				return []AddressTriple{}, answer

			case []AddressTriple:

				kademlia.RefreshClosest(answer, hash)
				if kademlia.identicalCalls > kademlia.identicalThreshold {
					fmt.Println("File not found (multiple consecutive same answers)")
					return kademlia.closest, ""
				} else {
					nextNode := kademlia.GetNextNode()
					if nextNode != nil {
						kademlia.goroutines++
						go kademlia.network.SendFindContact(ConvertToUDPAddr(*nextNode), hash, answerChannel)
					} else {
						fmt.Println("Thread ended")
						kademlia.goroutines--
					}
				}
			}

		default:
			if kademlia.goroutines == 0 {
				return kademlia.closest, ""
			}
			if kademlia.goroutines < kademlia.k {
				nextNode := kademlia.GetNextNode()
				if nextNode != nil {
					fmt.Println("Sending find contact message to node")
					kademlia.goroutines++
					go kademlia.network.SendFindContact(ConvertToUDPAddr(*nextNode), hash, answerChannel)
				}
			}
		}
	}
}

// Stores a file locally
func (kademlia *Kademlia) Store(data []byte, fileName string) {
	kademlia.fileChannel <- fileUtilsKademlia.Order{Action: fileUtilsKademlia.ADD, Name: fileName, Content: data}
}

// Goes through the list of closest contacts and returns the next node to ask
func (kademlia *Kademlia) GetNextNode() *AddressTriple {
	for index := range kademlia.closest {
		if kademlia.asked[kademlia.closest[index].Id] != true {
			kademlia.asked[kademlia.closest[index].Id] = true
			return &kademlia.closest[index]
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

	//return only K closest ones (remove the tail)
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
