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
	"os/exec"
	"bytes"
)

func main() {
	StartKademlia(true)
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
			fmt.Println("first node id : "+responseString)
			return kdmlib.AddressTriple{Ip: tripleElements[0], Port: tripleElements[1], Id: tripleElements[2]}
		}
	} else {
		fmt.Println(err)
	}
	return kdmlib.AddressTriple{}
}

func GetOutboundIP(isDocker bool) string {
	if !isDocker{
		conn, err := net.Dial("udp", "8.8.8.8:80")
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		localAddr := conn.LocalAddr().(*net.UDPAddr)

		return localAddr.IP.String()
	}else{
		cmd := exec.Command("hostname","-i")
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		if err!= nil {
			fmt.Println("error while executing 'hostname -I'")
			return ""
		}else{
			fmt.Println("")
		}
		ip := out.String()

		return strings.Replace(ip,"\n","",-1)
	}
}

func StartKademlia(isContainer bool) {
	nodeId := kdmlib.GenerateRandID(int64(rand.Intn(100)), 160)
	rt := kdmlib.CreateAllWorkersForRoutingTable(kdmlib.K, 160, 5, nodeId)

	chanPin, chanFile, fileMap := fileUtilsKademlia.CreateAndLaunchFileWorkers()

	port := "12000"
	ip := GetOutboundIP(isContainer)

	firstNode := SendAddress(ip, port, nodeId)
	fmt.Println(firstNode)
	if firstNode.Id != nodeId && firstNode.Id != "" {
		rt.GiveOrder(kdmlib.OrderForRoutingTable{kdmlib.ADD, firstNode, false})
		time.Sleep(time.Second)
	}
	nw := kdmlib.InitNetwork(port, ip, rt, nodeId, false, false, chanFile, chanPin, fileMap)
	kdm := kdmlib.NewKademliaInstance(nw, nodeId, kdmlib.ALPHA, kdmlib.K, rt, chanFile, fileMap)
	if firstNode.Id != nodeId && firstNode.Id != ""{
		contacts , _:= kdm.LookupAlgorithm(nodeId,kdmlib.ContactLookup)
		for _, j := range contacts{
			rt.GiveOrder(kdmlib.OrderForRoutingTable{kdmlib.ADD,j,false})
		}
	}
	fmt.Println("nodes in routing table:")
	fmt.Println(rt.FindKClosest(firstNode.Id))
	restApi.LaunchRestAPI(fileMap, chanFile, chanPin, *kdm)
}
