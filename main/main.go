package main

import (
	"Kademlia---P2P-DFS/kdmlib"
	"Kademlia---P2P-DFS/kdmlib/fileutils"
	"Kademlia---P2P-DFS/kdmlib/restApi"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"
	"net"
	"log"
)

func main() {
	StartKademlia()

}

func SendAddress(ip string, port string, id string) kdmlib.AddressTriple {
	req, err := http.Get("http://distributed.melittacloud.com:8080/add?ip=" + ip + "&port=" + port + "&id=" + id)
	if err == nil {
		fmt.Println("no error")
		body, err := ioutil.ReadAll(req.Body)

		if err == nil {
			responseString := string(body)
			fmt.Println(responseString)
			tripleElements := strings.Split(responseString, ",")
			return kdmlib.AddressTriple{Ip: tripleElements[0], Port: tripleElements[1], Id: tripleElements[2]}
		}
	} else {
		fmt.Println(err)
	}
	return kdmlib.AddressTriple{}
}

func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func StartKademlia() {
	nodeId := kdmlib.GenerateRandID(int64(rand.Intn(100)))
	rt := kdmlib.CreateAllWorkersForRoutingTable(kdmlib.K, kdmlib.IDLENGTH, 5, nodeId)

	chanPin, chanFile, fileMap := fileUtilsKademlia.CreateAndLaunchFileWorkers()

	port := "12000"
	ip := GetOutboundIP()

	firstNode := SendAddress(ip, port, nodeId)
	fmt.Println(firstNode)
	if firstNode.Id != nodeId && firstNode.Id != "" {
		rt.GiveOrder(kdmlib.OrderForRoutingTable{kdmlib.ADD, firstNode, false})
		time.Sleep(time.Second)
	}
	nw := kdmlib.InitNetwork(port, ip, rt, nodeId, false, chanFile, chanPin, fileMap)
	kdm := kdmlib.NewKademliaInstance(nw, nodeId, kdmlib.ALPHA, kdmlib.K, rt, chanFile, fileMap)
	restApi.LaunchRestAPI(fileMap, chanFile, chanPin, *kdm)
}
