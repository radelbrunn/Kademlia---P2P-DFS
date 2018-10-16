package main

import (
	"Kademlia---P2P-DFS/kdmlib"
	"Kademlia---P2P-DFS/kdmlib/fileutils"
	"Kademlia---P2P-DFS/kdmlib/restApi"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
)

func main() {
	StartKademlia()
}

func SendAddress(ip string, port string, id string) kdmlib.AddressTriple {
	req, err := http.Get("http://distributed.melittacloud.com:8080/add?ip=" + ip + "&port=" + port + "&id=" + id)
	if err != nil {
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			responseString := string(body)
			tripleElements := strings.Split(responseString, ",")
			return kdmlib.AddressTriple{Ip: tripleElements[0], Port: tripleElements[1], Id: tripleElements[2]}
		}
	}
	return kdmlib.AddressTriple{}
}

func StartKademlia() {
	nodeId := kdmlib.GenerateRandID(int64(rand.Intn(100)))
	rt := kdmlib.CreateAllWorkersForRoutingTable(kdmlib.K, kdmlib.IDLENGTH, 5, nodeId)
	chanPin, chanFile, fileMap := fileUtilsKademlia.CreateAndLaunchFileWorkers()
	port := "12000"
	ip := "127.0.0.1"
	firstNode := SendAddress(ip, port, nodeId)
	if firstNode.Id != nodeId {
		rt.GiveOrder(kdmlib.OrderForRoutingTable{kdmlib.ADD, firstNode, false})
	}
	nw := kdmlib.InitNetwork(port, ip, rt, nodeId, false)
	nfn := kdmlib.InitFileNetwork(ip, port)
	kdm := kdmlib.NewKademliaInstance(nw, nodeId, kdmlib.ALPHA, kdmlib.K, rt, nfn)
	restApi.LaunchRestAPI(fileMap, chanFile, chanPin, *kdm)
}
