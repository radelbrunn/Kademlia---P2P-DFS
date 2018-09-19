package main

import (
	"Kademlia---P2P-DFS/d7024e"
	"fmt"
)

func main() {
	contact := d7024e.NewContact(d7024e.NewKademliaID("FFFFFFFF00000000000000000000000000000000"), "localhost:8000")
	fmt.Print(contact.String())
}
