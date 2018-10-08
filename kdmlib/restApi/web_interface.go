package restApi

import (
	"log"
	"net/http"
	"github.com/gorilla/mux"
	"bytes"
	"strings"
	"fmt"
	"io"
	filesUtils "Kademlia---P2P-DFS/kdmlib/fileutils"
)

type RestDependencies struct {
	Map filesUtils.FileMap

}

// our main function
func LaunchRestAPI(fileMap filesUtils.FileMap) {
	router := mux.NewRouter()
	dependencies := RestDependencies{fileMap}
	router.HandleFunc("/{name}",dependencies.getFile).Methods("GET")
	router.HandleFunc("/{name}",ReceiveFile).Methods("POST")

	log.Fatal(http.ListenAndServe(":8000", router))
}

func (dependencies RestDependencies)getFile (w http.ResponseWriter, r *http.Request) {
	if dependencies.Map.IsPresent(mux.Vars(r)["name"]) {
		//TODO send Back the file
	}else{
		//TODO try to get the file from another node
	}



}

func ReceiveFile(w http.ResponseWriter, r *http.Request) {
	var Buf bytes.Buffer
	// in your case file would be fileupload
	file, header, err := r.FormFile("file")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	name := strings.Split(header.Filename, ".")
	fmt.Printf("File name %s\n", name[0])
	// Copy the file data to my buffer
	io.Copy(&Buf, file)
	// do something with the contents...
	// I normally have a struct defined and unmarshal into a struct, but this will
	// work as an example
	contents := Buf.String()
	fmt.Println(contents)
	// I reset the buffer in case I want to use it again
	// reduces memory allocations in more intense projects
	Buf.Reset()
	// do something else
	// etc write header
	return
}