package kdmlib

import "fmt"

type Kademlia struct {
	closest []AddressTriple
	asked   map[AddressTriple]bool
	nodeId  string
	rt      RoutingTable
	network Network
	alpha   int
	k       int
}

func NewKademliaInstance(nw *Network, nodeId string, alpha int, k int) *Kademlia {
	kademlia := &Kademlia{}
	kademlia.network = *nw
	kademlia.asked = make(map[AddressTriple]bool)
	kademlia.nodeId = GenerateRandID()
	kademlia.rt = CreateAllWorkersForRoutingTable(k, 160, 5, nodeId)
	kademlia.alpha = alpha
	kademlia.k = k
	return kademlia
}

func (kademlia *Kademlia) FindNextNodeToAsk() (nextContact *AddressTriple, success bool) {
	for i := range kademlia.closest {
		if kademlia.asked[kademlia.closest[i]] != true {
			kademlia.asked[kademlia.closest[i]] = true
			nextContact = &kademlia.closest[i]
			success = true
			return
		}
	}
	nextContact = nil
	success = false
	return
}

func (kademlia *Kademlia) LookupContact(target *AddressTriple) {
	kademlia.closest = []AddressTriple{}

	for _, e := range kademlia.rt.FindKClosest(target.Id) {
		kademlia.closest = append(kademlia.closest, e.Triple)
	}

	for i := 0; i < kademlia.alpha && i < len(kademlia.closest); i++ {
		fmt.Println("Sending find contact message")
		//TODO: send request to Raviljs Find Contact function
		//kademlia.network.SendFindContactMessage(kademlia.closest[i])

		kademlia.asked[kademlia.closest[i]] = true
	}

	//TODO: listen to response channel and call HandleChannelResponse func
}

func (kademlia *Kademlia) HandleChannelResponse(answer interface{}) (answerContacts []AddressTriple, dataAnswer string) {
	switch answer := answer.(type) {
	case []AddressTriple:
		return answer, ""
	case string:
		return nil, answer
	}

	return nil, ""
}

func (kademlia *Kademlia) LookupData(hash string) {
	// TODO
}

func (kademlia *Kademlia) Store(data []byte) {
	// TODO
}
