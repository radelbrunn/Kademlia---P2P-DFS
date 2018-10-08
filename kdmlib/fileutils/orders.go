package fileUtilsKademlia

const (
	ADD    = 0
	REMOVE = 1
)

type Order struct {
	Action  int
	Name    string
	Content []byte
}
