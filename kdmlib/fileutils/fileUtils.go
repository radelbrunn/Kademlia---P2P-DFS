package fileUtilsKademlia

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
	"path/filepath"
)

const fileDirectory = "./.files" + string(os.PathSeparator)

type pinnedFilesStruct struct {
	PinnedFiles map[string]bool
	lock        *sync.Mutex
}

type FileMap struct {
	MapPresent map[string]bool
	lock       *sync.Mutex
}

//creates a map and returns its pointer
func createPinnedFileList() *pinnedFilesStruct {
	pinnedFiles := pinnedFilesStruct{make(map[string]bool), &sync.Mutex{}}
	return &pinnedFiles
}

//checks if the file is pinned or not
func checkIfInList(pinnedFiles map[string]bool, name string) bool {
	isPresent, _ := pinnedFiles[name]
	return isPresent
}

//creates a directory to put files in
func createFilesDirectory() {
	os.Mkdir(fileDirectory, 0755)
}

//add or removes files from the node
func fileHandler(order Order, fileMap FileMap) {
	if order.Action == ADD {

		fileMap.set(order.Name, true)
		//checks if the file is already present
		if _, err := os.Stat(fileDirectory + order.Name); os.IsNotExist(err) {
			err := ioutil.WriteFile(fileDirectory+string(os.PathSeparator)+order.Name, order.Content, 0644)
			if err != nil {
				fmt.Println("something went wrong while creating file " + order.Name)
			}
			fmt.Println("no update")
		} else {
			fmt.Println("update")
			//only updates the file's modification date if it is already present
			updateFile(order.Name)
		}
	} else if order.Action == REMOVE {
		err := os.Remove(fileDirectory + string(os.PathSeparator) + order.Name)
		fileMap.set(order.Name, false)
		if err != nil {
			fmt.Println("something went wrong while removing file " + order.Name)
		}
	}
}

// create the fileDirectory directory and populates it according to incoming orders
func fileHandlerWorker(orders chan Order, fileMap FileMap) {
	createFilesDirectory()
	for {
		fileHandler(<-orders, fileMap)
	}
}

func (f FileMap) set(name string, isPresent bool) {
	f.lock.Lock()
	f.MapPresent[name] = isPresent
	f.lock.Unlock()
}

func (f FileMap) IsPresent(name string) bool {
	f.lock.Lock()
	isPresent := f.MapPresent[name]
	f.lock.Unlock()
	return isPresent
}

func populateFileMap() map[string]bool {
	filesMap := make(map[string]bool)
	dir , _ := filepath.Abs(filepath.Dir(os.Args[0]))
	fmt.Println(dir +fileDirectory)
	files, err := ioutil.ReadDir(fileDirectory)
	if err != nil {
		log.Fatal(err)
	}

	for j, f := range files {
		fmt.Println(j,":",f.Name())
		if !f.ModTime().Before(time.Now().Add(-time.Hour * 25)) {
			if !f.IsDir() {
				filesMap[f.Name()] = true
			}
		}
	}
	return filesMap
}

//pin or unpin a file
func pinFile(pinnedFiles *pinnedFilesStruct, ordersFromchan Order) {
	pinnedFiles.lock.Lock()
	if ordersFromchan.Action == ADD {
		if !checkIfInList(pinnedFiles.PinnedFiles, ordersFromchan.Name) {
			pinnedFiles.PinnedFiles[ordersFromchan.Name] = true
		}
	} else if ordersFromchan.Action == REMOVE {
		if checkIfInList(pinnedFiles.PinnedFiles, ordersFromchan.Name) {
			pinnedFiles.PinnedFiles[ordersFromchan.Name] = false
		}
	}
	pinnedFiles.lock.Unlock()
}

//reads from the channel and pin or unpin according to the order

func pinner(orders chan Order, pinnedFiles *pinnedFilesStruct) {
	for ordersFromchan := range orders {
		pinFile(pinnedFiles, ordersFromchan)
	}
}

//remove old files that are more than 25 hours old and not in the pinned list
func removeOldFiles(pinnedFiles *pinnedFilesStruct) {
	files, err := ioutil.ReadDir(fileDirectory)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		if f.ModTime().Before(time.Now().Add(-time.Hour * 25)) {
			pinnedFiles.lock.Lock()
			if !f.IsDir() && !checkIfInList(pinnedFiles.PinnedFiles, f.Name()) {
				fmt.Println(f.Name(), " : file too old, thus removing it")
				os.Remove(fileDirectory + f.Name())
			}
			pinnedFiles.lock.Unlock()
		}
	}
}

//wakes up once every hour to remove all the passed files not in the pinned list

func cleaner(pinnedFiles *pinnedFilesStruct) {
	for {
		removeOldFiles(pinnedFiles)
		time.Sleep(time.Hour)
	}
}

//reads file from os and returns a byte slice.
// Can be used to check if a file is present
func ReadFileFromOS(name string) []byte {
	dat, err := ioutil.ReadFile(fileDirectory + name)
	if err != nil {
		fmt.Println(err)
		return nil
	} else {
		return dat
	}
}

//update last modified date to now

func updateFile(name string) {
	os.Chtimes(fileDirectory+name, time.Now(), time.Now())
}

//creates all the workers needed to take care of the files
// and returns the channels to communicate with the workers
// 1st value is the channel for the pinner and the second one
// is the channel for the fileHandler
func CreateAndLaunchFileWorkers() (chan Order, chan Order, FileMap) {
	createFilesDirectory()
	channelForPinner := make(chan Order, 1000)
	channelForFileHandler := make(chan Order, 1000)
	pinnedFiles := createPinnedFileList()

	fileMap := FileMap{populateFileMap(), &sync.Mutex{}}
	fmt.Println("filemap :")
	for i,j := range fileMap.MapPresent {
		fmt.Println(i,":",j)
	}
	fmt.Println("end of filemap")
	go cleaner(pinnedFiles)
	go pinner(channelForPinner, pinnedFiles)
	go fileHandlerWorker(channelForFileHandler, fileMap)

	return channelForPinner, channelForFileHandler, fileMap
}
