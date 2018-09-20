package kdmlib

type Kademlia struct {
	closestContacts []AddressTriple
	node_id         string
	rtch            chan OrderForRoutingTable
	network         Network
	alpha           int
	k               int
}

func NewKademliaInstance(nw *Network, node_id string, alpha int, k int) *Kademlia {
	kademlia := &Kademlia{}
	kademlia.network = *nw
	kademlia.node_id = "RANDOM"
	kademlia.alpha = alpha
	kademlia.k = k
	return kademlia
}

func (kademlia *Kademlia) LookupContact(target *AddressTriple) {
	//rt, ch := CreateAllWorkersForRoutingTable(kademlia.k, 160, 5, kademlia.node_id)
	//kademlia.closestContacts = NewContactCandidates()

	/*
		for i := 0; i < kademlia.alpha; i++ {
			//TODO
			fmt.Println("Send Find Contact request")
		}
	*/
}

//hash is the hashed filename???
func (kademlia *Kademlia) LookupData(hash string) {
	/*
		kademlia.closestContacts = NewContactCandidates()

		//TODO Convert hash to a *KademliaID
		//kademlia.closestContacts.Append(kademlia.rt.FindClosestContacts(hash, kademlia.alpha))

		for i := 0; i < kademlia.alpha; i++ {
			//TODO
			fmt.Println("Send Find Data")
		}
	*/
}

func (kademlia *Kademlia) Store(data []byte) {
	// TODO
}
