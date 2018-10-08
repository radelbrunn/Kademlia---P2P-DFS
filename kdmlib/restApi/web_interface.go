package restApi

import (
	"net/http"
	"github.com/gorilla/mux"
	filesUtils "Kademlia---P2P-DFS/kdmlib/fileutils"
	"log"
	"strconv"
)

type RestDependencies struct {
	Map         filesUtils.FileMap
	FileChannel chan filesUtils.Order
	Pinning     chan filesUtils.Order
}

// main function
func LaunchRestAPI(fileMap filesUtils.FileMap, fileChannel chan filesUtils.Order, pinningChannel chan filesUtils.Order) {
	router := mux.NewRouter()
	dependencies := RestDependencies{fileMap, fileChannel, pinningChannel}
	router.HandleFunc("/{name}", dependencies.getFile).Methods("GET")
	router.HandleFunc("/", dependencies.receiveFile).Methods("POST")
	router.HandleFunc("/pin/{name}", dependencies.pin).Methods("PUT")
	router.HandleFunc("/unpin/{name", dependencies.unpin).Methods("PUT")

	log.Fatal(http.ListenAndServe(":8000", router))
}

func (dependencies RestDependencies) getFile(w http.ResponseWriter, r *http.Request) {
	if dependencies.Map.IsPresent(mux.Vars(r)["name"]) {
		w.WriteHeader(200)
		w.Header().Set("Content-Disposition", "attachment; filename="+mux.Vars(r)["name"])
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		value := filesUtils.ReadFileFromOS(mux.Vars(r)["name"])
		w.Header().Set("Content-Length", strconv.Itoa(len(value)))

		w.Write(value)
	} else {
		//TODO try to get the file from another node , send back 404 if not found, 200 and file if found

	}
}

func (dependencies RestDependencies) receiveFile(w http.ResponseWriter, r *http.Request) {
	//TODO read file into a buffer, send order to channel, and launch the right function to send the file
	// to the other nodes and compute 160 bits hash

}

func (dependencies RestDependencies) pin(w http.ResponseWriter, r *http.Request) {
	dependencies.Pinning <- filesUtils.Order{filesUtils.ADD, mux.Vars(r)["name"], nil}
	w.WriteHeader(200)
	w.Write(nil)
}

func (dependencies RestDependencies) unpin(w http.ResponseWriter, r *http.Request) {
	dependencies.Pinning <- filesUtils.Order{filesUtils.REMOVE, mux.Vars(r)["name"], nil}
	//TODO add function to locate the nodes who have the file
	w.WriteHeader(200)
	w.Write(nil)
}