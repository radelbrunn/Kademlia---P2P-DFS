package kdmlib

import (
	"sync"
)

type Node struct {
	rt                     *routingTableAndCache
	mux                    *sync.Mutex
	republishTimeSeconds   int
	republishRandomSeconds int
}
