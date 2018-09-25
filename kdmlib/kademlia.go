package kdmlib

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

	//tableClosest = kademlia.rt.FindKClosest(target.Id)

	for i := 0; i < kademlia.alpha && i < len(kademlia.closest); i++ {

	}

	//kademlia.closest = append(kademlia.closest, kademlia.rt.FindKClosest(target.Id))

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
