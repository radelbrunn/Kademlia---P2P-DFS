package fileUtilsKademlia

import (
	"fmt"
	"time"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

//checks if the file is pinned or not
type PinnedFilesStruct struct {
	pinnedFiles map[string]bool
	lock *sync.Mutex
}

func CreatePinnedFileList() *PinnedFilesStruct{
	return &PinnedFilesStruct{make(map[string]bool),&sync.Mutex{}}
}

func checkIfInList(pinnedFiles map[string]bool, name string) bool {
	isPresent , _ := pinnedFiles[name]
	return isPresent
}



func createFilesDirectory(){
	os.Mkdir("./files",0644)
}

//add or removes files from the node
func fileHandler(order Order) {
	if order.action == ADD {
		err := ioutil.WriteFile(".files/"+order.name, order.content, 0644)
		if err != nil {
			fmt.Println("something went wrong while creating file " + order.name)
		}
	} else if order.action == REMOVE {
		err := os.Remove(".files/"+order.name)
		if err != nil {
			fmt.Println("something went wrong while removing file " + order.name)
		}
	}
}

// create the ".file/" directory and populates it according to incoming orders
func FileHandlerWorker(orders chan Order){
	createFilesDirectory()
	for {
		fileHandler(<-orders)
	}
}


//reads from the channel and pin or unpin according to the order

func Pinner(orders <-chan Order, pinnedFiles *PinnedFilesStruct) {
	for ordersFromchan := range orders {
		pinnedFiles.lock.Lock()
		if ordersFromchan.action == ADD {
			if !checkIfInList(pinnedFiles.pinnedFiles, ordersFromchan.name)  {
				pinnedFiles.pinnedFiles[ordersFromchan.name]=true
			}
		} else if ordersFromchan.action == REMOVE {
			if checkIfInList(pinnedFiles.pinnedFiles, ordersFromchan.name) {
				pinnedFiles.pinnedFiles[ordersFromchan.name]=false
			}
		}
		pinnedFiles.lock.Unlock()
	}
}

//wakes up once every hour to remove all the passed files not in the pinned list

func Cleaner(pinnedFiles *PinnedFilesStruct) {
	for {
		files, err := ioutil.ReadDir("./.files/")
		if err != nil {
			log.Fatal(err)
		}

		for _, f := range files {
			if f.ModTime().Before(time.Now().Add(-time.Hour * 25)) {
				pinnedFiles.lock.Lock()
				if !f.IsDir() && checkIfInList(pinnedFiles.pinnedFiles, f.Name()) {
					fmt.Println(f.Name(), " : file too old, thus removing it")
					os.Remove(f.Name())
				}
				pinnedFiles.lock.Unlock()
			}
		}
		time.Sleep(time.Hour)
	}
}

//update last modified date to now

func UpdateFile(name string) {
	os.Chtimes(name, time.Now(), time.Now())
}
