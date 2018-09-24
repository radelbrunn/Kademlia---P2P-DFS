package fileUtilsKademlia

import (
	"fmt"
	"container/list"
	"time"
	"io/ioutil"
	"log"
	"os"
)

//checks if the file is pinned or not

func checkIfInList(list *list.List, name string) *list.Element {
	for str := list.Front(); str != nil; str = str.Next() {
		if str.Value == name {
			return str
		}
	}
	return nil
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


//reads from the channel and pin or unpin arccording to the order

func Pinner(orders <-chan Order, pinnedFiles *list.List) {
	for ordersFromchan := range orders {
		if ordersFromchan.action == ADD {
			if checkIfInList(pinnedFiles, ordersFromchan.name) == nil {
				pinnedFiles.PushBack(ordersFromchan.name)
			}

		} else if ordersFromchan.action == REMOVE {
			if file := checkIfInList(pinnedFiles, ordersFromchan.name); file != nil {
				pinnedFiles.Remove(file)
			}
		}
	}
}

//wakes up once every hour to remove all the passed files not in the pinned list

func Cleaner(pinnedFiles *list.List) {
	for {
		files, err := ioutil.ReadDir("./.files/")
		if err != nil {
			log.Fatal(err)
		}

		for _, f := range files {
			if f.ModTime().Before(time.Now().Add(-time.Hour * 25)) {
				if !f.IsDir() && checkIfInList(pinnedFiles, f.Name()) == nil {
					fmt.Println(f.Name(), " : file too old, thus removing it")
					os.Remove(f.Name())
				}
			}
		}
		time.Sleep(time.Hour)
	}
}

//update last modified date to now

func UpdateFile(name string) {
	os.Chtimes(name, time.Now(), time.Now())
}
