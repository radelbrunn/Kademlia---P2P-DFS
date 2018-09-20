package kdmlib

import (
	"math/rand"
	"strconv"
	"time"
)

const (
	K        = 20
	ALPHA    = 3
	IDLENGTH = 160
)

func GenerateRandID() string {

	id := ""
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < IDLENGTH; i++ {
		id += strconv.Itoa(rand.Intn(2))
	}

	return id
}
