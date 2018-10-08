package fileUtilsKademlia

import (
	"testing"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
	"fmt"
	"sync"
)

func TestReadFileExist(t *testing.T) {
	os.Mkdir(".files/",0755)
	ioutil.WriteFile(".files"+string(os.PathSeparator)+"testFile",[]byte("hello World"), 0644)
	dat := ReadFileFromOS("testFile")
	result := string(dat)
	if result!="hello World"{
		t.Error("wrong file content or content")
	}
	os.Remove(".files"+string(filepath.Separator)+"testFile")
	os.Remove(".files")
}

func TestUpdateFile(t *testing.T){
	ioutil.WriteFile("testFile",[]byte("hello world"),0644)
	fi, _ := os.Stat("testFile")
	modTime := fi.ModTime()
	updateFile("testFile")
	fi, _ = os.Stat("testFile")
	if fi.ModTime().Before(modTime){
		t.Error("file not updated")
	}
	os.Remove("testFile")
}

func TestReadFileExistNot(t *testing.T) {
	os.Mkdir(".files/",0755)
	dat := ReadFileFromOS("testFile")
	if dat!=nil{
		t.Error("wrong file content or content")
	}
	os.Remove(".files"+string(filepath.Separator)+"testFile")
	os.Remove(".files")
}

func TestRemoveOldFilesNotPinnedTooOld(t *testing.T){
	os.Mkdir(".files/",0755)

	pinnedFiles := createPinnedFileList()
	ioutil.WriteFile(fileDirectory+string(filepath.Separator)+"testFile",[]byte("hello world"),0644)

	os.Chtimes(".files"+string(filepath.Separator)+
		"testFile",time.Now().Add(-time.Hour*25),time.Now().Add(-time.Hour*25))

	removeOldFiles(pinnedFiles)

	files, _ := ioutil.ReadDir(fileDirectory)

	for _, f := range files{
		fmt.Println(f.Name())
		if !f.IsDir(){
			t.Error("file not removed but should be removed")
		}
	}

	os.Remove(".files"+string(filepath.Separator)+"testFile")
	os.Remove(".files")
}

func TestRemoveOldFilesNotPinnedOK(t *testing.T){
	os.Mkdir(".files/",0755)

	pinnedFiles := createPinnedFileList()
	ioutil.WriteFile(fileDirectory+string(filepath.Separator)+"testFile",[]byte("hello world"),0644)

	removeOldFiles(pinnedFiles)

	files, _ := ioutil.ReadDir(fileDirectory)
	isremoved := true
	for _, f := range files{
		fmt.Println(f.Name())
		if f.Name()=="testFile"{
			isremoved=false
		}
	}
	if isremoved{
		t.Error("file not removed but should be removed")
	}

	os.Remove(".files"+string(filepath.Separator)+"testFile")
	os.Remove(".files")
}

func TestShouldNotRemovePinnedFile(t *testing.T){
	os.Mkdir(".files/",0755)

	pinnedFiles := createPinnedFileList()
	ioutil.WriteFile(fileDirectory+string(filepath.Separator)+"testFile",[]byte("hello world"),0644)

	os.Chtimes(fileDirectory+string(os.PathSeparator)+"testFile",
		time.Now().Add(-time.Hour*25), time.Now().Add(-time.Hour*25))

	pinFile(pinnedFiles,Order{ADD,"testFile",nil})

	removeOldFiles(pinnedFiles)
	files, _ := ioutil.ReadDir(fileDirectory)
	isremoved := true
	for _, f := range files{
		fmt.Println(f.Name())
		if f.Name()=="testFile"{
			isremoved=false
		}
	}
	if isremoved{
		t.Error("file removed but should not be removed")
	}

	os.Remove(".files"+string(filepath.Separator)+"testFile")
	os.Remove(".files")
}

func TestShouldRemoveFile(t *testing.T){
	os.Mkdir(".files/",0755)

	pinnedFiles := createPinnedFileList()
	ioutil.WriteFile(fileDirectory+string(filepath.Separator)+"testFile",[]byte("hello world"),0644)

	os.Chtimes(fileDirectory+string(os.PathSeparator)+"testFile",
		time.Now().Add(-time.Hour*25), time.Now().Add(-time.Hour*25))
	removeOldFiles(pinnedFiles)

	files, _ := ioutil.ReadDir(fileDirectory)
	isremoved := true
	for _, f := range files{
		fmt.Println(f.Name())
		if f.Name()=="testFile"{
			isremoved=false
		}
	}
	if !isremoved{
		t.Error("file not removed but should be removed")
	}

	os.Remove(".files"+string(filepath.Separator)+"testFile")
	os.Remove(".files")
}



func TestCreateFile(t *testing.T){
	os.Mkdir(".files/",0755)
	fm := FileMap{make(map[string]bool),&sync.Mutex{}}

	fileHandler(Order{ADD,"testFile",[]byte("helloWorld")},fm)

	dat := ReadFileFromOS("testFile")

	fmt.Println(string(dat))
	if string(dat)!="helloWorld" {
		t.Error("file not written correctly")
	}

	os.Remove(".files"+string(filepath.Separator)+"testFile")
	os.Remove(".files")
}

func TestRemoveFile(t *testing.T){
	os.Mkdir(".files/",0755)
	fm := FileMap{make(map[string]bool),&sync.Mutex{}}
	fileHandler(Order{ADD,"testFile",[]byte("helloWorld")},fm)
	fileHandler(Order{REMOVE,"testFile",[]byte("helloWorld")},fm)

	files, _ := ioutil.ReadDir(fileDirectory)
	isremoved := true
	for _, f := range files{
		fmt.Println(f.Name())
		if f.Name()=="testFile"{
			isremoved=false
		}
	}
	if !isremoved{
		t.Error("file not removed but should be")
	}

	os.Remove(".files"+string(filepath.Separator)+"testFile")
	os.Remove(".files")
}


func TestUpdateFileFromHandler (t *testing.T){
	os.Mkdir(".files/",0755)
	fm := FileMap{make(map[string]bool),&sync.Mutex{}}

	fileHandler(Order{ADD,"testFile",[]byte("helloWorld")},fm)
	timenow := time.Now()
	fileHandler(Order{ADD,"testFile",[]byte("helloWorld")},fm)
	fi , _ := os.Stat(fileDirectory+string(os.PathSeparator)+"testFile")
	if !timenow.Before(fi.ModTime()){
		t.Error("time should be updated and is not")
	}


	os.Remove(".files"+string(filepath.Separator)+"testFile")
	os.Remove(".files")
}

func TestPinRemove(t *testing.T){
	pinnedFiles := createPinnedFileList()
	pinFile(pinnedFiles,Order{ADD,"toto",nil})
	pinFile(pinnedFiles,Order{REMOVE,"toto",nil})
	if pinnedFiles.pinnedFiles["toto"]{
		t.Error("entry not removed")
	}
}


