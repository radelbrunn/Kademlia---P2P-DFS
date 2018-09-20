package kdmlib

import ()

type Network struct {
}

func NewNetwork(ip string, port int) *Network {

	network := &Network{}

	return network
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
