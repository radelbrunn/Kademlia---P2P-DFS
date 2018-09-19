package d7024e

type Kademlia struct {
	rt      *RoutingTable
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
	// TODO
}

func (kademlia *Kademlia) LookupData(hash string) {
	// TODO
}

func (kademlia *Kademlia) Store(data []byte) {
	// TODO
}
