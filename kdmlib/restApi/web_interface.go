package restApi

import (
	"log"
	"net/http"
	"github.com/gorilla/mux"
)

// our main function
func LaunchRestAPI() {
	router := mux.NewRouter()
	router.HandleFunc("/{name}",getFile).Methods("GET")
	router.HandleFunc("/{name}",uploadFile).Methods("POST")

	log.Fatal(http.ListenAndServe(":8000", router))
}

func getFile(w http.ResponseWriter, r *http.Request) {
	//TODO refactor to use the routing table and the network function
}

func uploadFile(w http.ResponseWriter,r *http.Request){
	//TODO refactor to use the routing table and the network function
}