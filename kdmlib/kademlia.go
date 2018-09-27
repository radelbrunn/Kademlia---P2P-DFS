package kdmlib

import (
	"fmt"
	"reflect"
)

type Kademlia struct {
	closest            []AddressTriple
	asked              map[AddressTriple]bool
	nodeId             string
	rt                 RoutingTable
	network            Network
	alpha              int
	k                  int
	goroutines         int
	identicalCalls     int
	identicalThreshold int
}

func NewKademliaInstance(nw *Network, nodeId string, alpha int, k int, rt RoutingTable) *Kademlia {
	kademlia := &Kademlia{}
	kademlia.network = *nw
	kademlia.asked = make(map[AddressTriple]bool)
	kademlia.nodeId = nodeId
	kademlia.rt = rt
	kademlia.alpha = alpha
	kademlia.k = k
	kademlia.goroutines = 0
	kademlia.identicalCalls = 0
	kademlia.identicalThreshold = alpha

	return kademlia
}

func (kademlia *Kademlia) GetNextNode() *AddressTriple {
	for index := range kademlia.closest {
		if kademlia.asked[kademlia.closest[index]] != true {
			kademlia.asked[kademlia.closest[index]] = true
			return &kademlia.closest[index]
		}
	}
	return nil
}

func (kademlia *Kademlia) LookupContact(target *AddressTriple) []AddressTriple {
	answerChannel := make(chan interface{}, kademlia.alpha)
	kademlia.closest = []AddressTriple{}

	for _, e := range kademlia.rt.FindKClosest(target.Id) {
		kademlia.closest = append(kademlia.closest, e.Triple)
	}

	for i := 0; i < kademlia.alpha && i < len(kademlia.closest); i++ {
		fmt.Println("Sending find contact message")
		kademlia.goroutines++
		kademlia.network.SendFindContactRequest(ConvertToUDPAddr(kademlia.closest[i]), target.Id, answerChannel)

		kademlia.asked[kademlia.closest[i]] = true
	}

	for {
		select {
		case answer := <-answerChannel:
			if reflect.TypeOf(answer) == reflect.TypeOf(kademlia.closest) {
				//TODO update closest contacts
				//TODO: handle exit on same occurrence after update
				nextNode := kademlia.GetNextNode()
				if nextNode != nil {
					kademlia.network.SendFindContactRequest(ConvertToUDPAddr(*nextNode), target.Id, answerChannel)
				}
			}
		default:
			if kademlia.goroutines == 0 {
				return kademlia.closest
			}
		}
	}
}

func (kademlia *Kademlia) LookupData(hash string) string {
	answerChannel := make(chan interface{}, kademlia.alpha)
	kademlia.closest = []AddressTriple{}

	for _, e := range kademlia.rt.FindKClosest(hash) {
		kademlia.closest = append(kademlia.closest, e.Triple)
	}

	for i := 0; i < kademlia.alpha && i < len(kademlia.closest); i++ {
		fmt.Println("Sending find data message")
		kademlia.network.SendFindDataRequest(ConvertToUDPAddr(kademlia.closest[i]), hash, answerChannel)

		kademlia.asked[kademlia.closest[i]] = true
	}

	for {
		select {
		case answer := <-answerChannel:
			switch answer := answer.(type) {
			case string:
				return answer

			case []AddressTriple:
				//TODO: update closest
				//TODO: handle exit on same occurrence after update
				nextNode := kademlia.GetNextNode()
				if nextNode != nil {
					kademlia.network.SendFindDataRequest(ConvertToUDPAddr(*nextNode), hash, answerChannel)
				}
			}

		default:
			if kademlia.goroutines == 0 {
				return ""
			}
		}
	}

}

func (kademlia *Kademlia) RefreshClosest(newContacts []AddressTriple, target string) {
	identicalList := true
	//newList := []AddressTriple{}
	for i := range kademlia.closest {
		for j := range newContacts {
			if kademlia.closest[i].Id == newContacts[j].Id {

			} else {
				identicalList = false
			}
		}
	}

	if identicalList {
		kademlia.identicalCalls++
	} else {
		kademlia.identicalCalls = 0
	}
}

func (kademlia *Kademlia) Store(data []byte) {
	// TODO Use Jeremys Library
}
