package restApi

import (
	filesUtils "Kademlia---P2P-DFS/kdmlib/fileutils"
	"crypto/sha1"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"Kademlia---P2P-DFS/kdmlib"
)

type RestDependencies struct {
	Map         filesUtils.FileMap
	FileChannel chan filesUtils.Order
	Pinning     chan filesUtils.Order
	kdm 		kdmlib.Kademlia
}

// main function
func LaunchRestAPI(fileMap filesUtils.FileMap, fileChannel chan filesUtils.Order, pinningChannel chan filesUtils.Order,kdm kdmlib.Kademlia) {
	router := mux.NewRouter()
	dependencies := RestDependencies{fileMap, fileChannel, pinningChannel,kdm}
	router.HandleFunc("/{name}", dependencies.getFile).Methods("GET")
	router.HandleFunc("/", dependencies.receiveFile).Methods("POST")
	router.HandleFunc("/pin/{name}", dependencies.pin).Methods("PUT")
	router.HandleFunc("/unpin/{name}", dependencies.unpin).Methods("PUT")

	log.Fatal(http.ListenAndServe(":8000", router))
}

func (dependencies RestDependencies) getFile(w http.ResponseWriter, r *http.Request) {
	args := mux.Vars(r)["name"]
	if dependencies.Map.IsPresent(args) {
		w.WriteHeader(200)
		w.Header().Set("Content-Disposition", "attachment; filename="+mux.Vars(r)["name"])
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		value := filesUtils.ReadFileFromOS(mux.Vars(r)["name"])
		w.Header().Set("Content-Length", strconv.Itoa(len(value)))
		w.Write(value)
	} else {

		if result := dependencies.kdm.LookupData(args,false);result!=nil{
			w.WriteHeader(200)
			w.Write([]byte("true"))
		}else{
			w.WriteHeader(404)
			w.Write([]byte("File not found"))
		}
	}
}

func (dependencies RestDependencies) receiveFile(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	hash := sha1.Sum(b)
	stringHash := string(hash[:])

	dependencies.FileChannel <- filesUtils.Order{filesUtils.ADD, stringHash, b}
	go dependencies.kdm.StoreData(stringHash,b,false)

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

	if dependencies.Map.IsPresent(name){
		dependencies.Pinning <- filesUtils.Order{Action: filesUtils.ADD, Name: name}
		w.WriteHeader(200)
		w.Write([]byte("file pinned"))
	}else{
		w.WriteHeader(404)
		w.Write([]byte("file doesn't exist on this node"))
	}
}

func (dependencies RestDependencies) unpin(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	if dependencies.Map.IsPresent(name){
		dependencies.Pinning <- filesUtils.Order{Action: filesUtils.REMOVE, Name: name}
		w.WriteHeader(200)
		w.Write([]byte("file unpinned"))
	}else{
		w.WriteHeader(404)
		w.Write([]byte("file doesn't exist on this node"))
	}
}
