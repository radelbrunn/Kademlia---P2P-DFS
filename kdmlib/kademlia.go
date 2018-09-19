package kdmlib

import "fmt"

type Kademlia struct {
	closestContacts ContactCandidates
	rt              *RoutingTable
	//rt				*RoutingTableAndCache
	network Network
	alpha   int
	k       int
}

func NewKademliaInstance(nw *Network) *Kademlia {
	kademlia := &Kademlia{}
	kademlia.network = *nw
	kademlia.alpha = 3
	kademlia.k = 20
	return kademlia
}

func (kademlia *Kademlia) LookupContact(target *Contact) {
	kademlia.closestContacts = NewContactCandidates()
	kademlia.closestContacts.Append(kademlia.rt.FindClosestContacts(target.ID, kademlia.alpha))

	for i := 0; i < kademlia.alpha; i++ {
		//TODO
		fmt.Println("Send Find Contact request")
	}
}

//hash is the hashed filename???
func (kademlia *Kademlia) LookupData(hash string) {
	kademlia.closestContacts = NewContactCandidates()

	//TODO Convert hash to a *KademliaID
	//kademlia.closestContacts.Append(kademlia.rt.FindClosestContacts(hash, kademlia.alpha))

	for i := 0; i < kademlia.alpha; i++ {
		//TODO
		fmt.Println("Send Find Data")
	}
}

func (kademlia *Kademlia) Store(data []byte) {
	// TODO
}
