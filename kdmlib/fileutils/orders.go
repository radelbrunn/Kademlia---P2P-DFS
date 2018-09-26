package fileUtilsKademlia
const (
	ADD  = 0
	REMOVE  = 1
)

type Order struct {
	action int
	name   string
	content []byte
}


