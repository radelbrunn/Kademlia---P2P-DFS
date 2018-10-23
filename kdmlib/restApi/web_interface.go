package restApi

import (
	"Kademlia---P2P-DFS/kdmlib"
	filesUtils "Kademlia---P2P-DFS/kdmlib/fileutils"
	"crypto/sha1"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type RestDependencies struct {
	Map         filesUtils.FileMap
	FileChannel chan filesUtils.Order
	Pinning     chan filesUtils.Order
	kdm         kdmlib.Kademlia
	myFiles     map[string]bool
	lock        *sync.Mutex
}

// main function
func LaunchRestAPI(fileMap filesUtils.FileMap, fileChannel chan filesUtils.Order, pinningChannel chan filesUtils.Order, kdm kdmlib.Kademlia) {
	router := mux.NewRouter()
	dependencies := RestDependencies{fileMap, fileChannel, pinningChannel, kdm, make(map[string]bool), &sync.Mutex{}}

	router.HandleFunc("/{name}", dependencies.getFile).Methods("GET")
	router.HandleFunc("/{name}", dependencies.deleteFile).Methods("DELETE")
	router.HandleFunc("/", dependencies.receiveFile).Methods("POST")
	router.HandleFunc("/pin/{name}", dependencies.pin).Methods("PUT")
	router.HandleFunc("/unpin/{name}", dependencies.unpin).Methods("PUT")

	log.Fatal(http.ListenAndServe(":8000", router))
}

func (dependencies RestDependencies) getFile(w http.ResponseWriter, r *http.Request) {
	args := mux.Vars(r)["name"]
	if dependencies.Map.IsPresent(args) {
		fmt.Println("isPresemt")
		w.WriteHeader(200)
		w.Header().Set("Content-Disposition", "attachment; filename="+args)
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		value := filesUtils.ReadFileFromOS(mux.Vars(r)["name"])
		w.Header().Set("Content-Length", strconv.Itoa(len(value)))
		w.Write(value)
	} else {

		if result := dependencies.kdm.LookupData(args); result != nil {
			w.WriteHeader(200)
			w.Write(result)
		} else {
			w.WriteHeader(404)
			w.Write([]byte("File not found"))
		}
	}
}

func (dependencies RestDependencies) deleteFile(w http.ResponseWriter, r *http.Request) {
	dependencies.lock.Lock()
	name := mux.Vars(r)["name"]
	isPresent := dependencies.myFiles[name]
	dependencies.lock.Unlock()
	if isPresent {
		dependencies.lock.Lock()
		dependencies.myFiles[name] = false
		dependencies.lock.Unlock()
		dependencies.FileChannel <- filesUtils.Order{filesUtils.REMOVE, name, nil}
		w.WriteHeader(200)
		w.Write([]byte("file deleted from this node"))
	} else {
		w.WriteHeader(404)
		w.Write([]byte("could not delete the file as the node is not the original publisher"))
	}
}

func (dependencies RestDependencies) receiveFile(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	args := r.URL.Query()
	hash := sha1.Sum(b)
	var stringHash string
	for _, i := range hash {
		j := int64(i)
		str2 := strconv.FormatInt(j, 2)
		for l := len(str2); l < 8; l++ {
			str2 = "0" + str2
		}
		//fmt.Println(str2)
		//fmt.Println(len(str2))
		stringHash = stringHash + str2
	}

	if !(args.Get("fromNetwork") == "true") {
		dependencies.updateOriginalFile(stringHash, true)
		go dependencies.kdm.StoreData(stringHash)
	}
	dependencies.FileChannel <- filesUtils.Order{filesUtils.ADD, stringHash, b}

	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("server error"))
	} else {
		w.WriteHeader(200)
		w.Write([]byte(stringHash))
	}
}

func (dependencies RestDependencies) pin(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	if dependencies.Map.IsPresent(name) {
		dependencies.Pinning <- filesUtils.Order{Action: filesUtils.ADD, Name: name}
		w.WriteHeader(200)
		w.Write([]byte("file pinned"))
	} else {
		w.WriteHeader(404)
		w.Write([]byte("file doesn't exist on this node"))
	}
}

func (dependencies RestDependencies) unpin(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	if dependencies.Map.IsPresent(name) {
		dependencies.Pinning <- filesUtils.Order{Action: filesUtils.REMOVE, Name: name}
		w.WriteHeader(200)
		w.Write([]byte("file unpinned"))
	} else {
		w.WriteHeader(404)
		w.Write([]byte("file doesn't exist on this node"))
	}
}

func (dependencies RestDependencies) republish() {
	for {
		dependencies.lock.Lock()
		for i, j := range dependencies.myFiles {
			if j {
				dependencies.kdm.StoreData(i)
			}
		}
		dependencies.lock.Unlock()
	}
	time.Sleep(time.Hour * 20)
}

func (dependencies RestDependencies) updateOriginalFile(fileName string, isPresent bool) {
	dependencies.lock.Lock()
	dependencies.myFiles[fileName] = isPresent
	dependencies.lock.Unlock()
}
