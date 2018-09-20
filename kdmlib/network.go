package kdmlib

type Network struct {
}

func Listen(ip string, port int) {
	// TODO
}

func (network *Network) SendPingMessage(contact *AddressTriple) {
	// TODO
}

func (network *Network) SendFindContactMessage(contact *AddressTriple) {
	// TODO
}

func (network *Network) SendFindDataMessage(hash string) {
	// TODO
}

func (network *Network) SendStoreMessage(data []byte) {
	// TODO
}
