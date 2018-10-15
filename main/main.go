package main

import (
	"Kademlia---P2P-DFS/kdmlib"
	"net/http"
	"io/ioutil"
	"strings"
	"math/rand"
	"fmt"
	"Kademlia---P2P-DFS/kdmlib/fileutils"
	"Kademlia---P2P-DFS/kdmlib/restApi"
)

func main() {
	StartKademlia()
}

func SendAddress(ip string,port string, id string) kdmlib.AddressTriple{
	req , err:= http.Get("http://distributed.melittacloud.com:8080/add?ip="+ip+"&port="+port+"&id="+id)
	if err == nil{
		fmt.Println("no error")
		body,err := ioutil.ReadAll(req.Body)

		if err== nil {
			responseString := string(body)
			fmt.Println(responseString)
			tripleElements := strings.Split(responseString,",")
			return kdmlib.AddressTriple{Ip: tripleElements[0], Port: tripleElements[1], Id: tripleElements[2]}
		}
	}else {
		fmt.Println(err)
	}
	return kdmlib.AddressTriple{}
}

func StartKademlia() {
	nodeId := kdmlib.GenerateRandID(int64(rand.Intn(100)))
	rt := kdmlib.CreateAllWorkersForRoutingTable(kdmlib.K, kdmlib.IDLENGTH, 5, nodeId)

	chanPin,chanFile,fileMap := fileUtilsKademlia.CreateAndLaunchFileWorkers()
	port := "12000"
	ip := "127.0.0.1"
	firstNode := SendAddress(ip,port,nodeId)
	fmt.Println(firstNode)
	if firstNode.Id!=nodeId && firstNode.Id!=""{
		rt.GiveOrder(kdmlib.OrderForRoutingTable{kdmlib.ADD,firstNode,false})
	}
	nw := kdmlib.InitNetwork(port, ip, rt, nodeId, false)
	kdm := kdmlib.NewKademliaInstance(nw, nodeId, kdmlib.ALPHA, kdmlib.K, rt)
	restApi.LaunchRestAPI(fileMap,chanFile,chanPin,*kdm)
}
