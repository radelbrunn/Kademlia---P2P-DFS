package kdmlib

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

const BUFFERSIZE = 1024

type fileNetwork struct {
	port        string
	ip          string
	conn        net.Listener
	packetsChan chan TCPPacketAndInfo
}
type TCPPacketAndInfo struct {
	connection net.Conn
	packet     []byte
}

func InitFileNetwork(ip string, port string) *fileNetwork {
	fileNetwork := &fileNetwork{}
	fileNetwork.port = port
	fileNetwork.ip = ip
	fileNetwork.packetsChan = make(chan TCPPacketAndInfo, 500)

	buffer := make([]byte, 4096)

	server, err := net.Listen("tcp", ip+":"+port)
	if err != nil {
		fmt.Println("Error listening: ", err)
		os.Exit(1)
	}
	fileNetwork.conn = server

	go fileNetwork.TCPServer(3, buffer)

	return fileNetwork
}
func (fileNetwork *fileNetwork) TCPServer(numberOfWorkers int, buffer []byte) {
	for i := 0; i < numberOfWorkers; i++ {
		go fileNetwork.ConnectionWorker()
	}
	defer fileNetwork.conn.Close()
	fmt.Println("Server started! Waiting for connections...")
	for {
		conn, err := fileNetwork.conn.Accept()
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
		fmt.Println("Client connected")

		fileNetwork.packetsChan <- TCPPacketAndInfo{connection: conn, packet: buffer}
	}
}

//reads from the channel and handles the packet
func (fileNetwork *fileNetwork) ConnectionWorker() {
	for toto := range fileNetwork.packetsChan {
		if fileNetwork.conn != nil {
			n, error := toto.connection.Read(toto.packet)
			if error != nil {
				fmt.Println("There is an error reading from connection", error.Error())
				return
			}
			filename := string(toto.packet[:n])
			SendFileToClient(toto.connection, filename)
		}
	}
}

func SendFileToClient(connection net.Conn, filename string) {
	fmt.Println("A client has connected!")
	defer connection.Close()
	file, err := os.Open("C:/Users/ReLaX/Desktop/" + filename) //just for testing
	if err != nil {
		fmt.Println(err)
		return
	}
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}
	fileSize := fillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
	fileName := fillString(fileInfo.Name(), 64)
	fmt.Println("Sending filename and filesize!")
	connection.Write([]byte(fileSize))
	connection.Write([]byte(fileName))
	sendBuffer := make([]byte, BUFFERSIZE)
	fmt.Println("Start sending file!")
	for {
		_, err = file.Read(sendBuffer)
		if err == io.EOF {
			break
		}
		connection.Write(sendBuffer)
	}
	fmt.Println("File has been sent, closing connection!")
	return
}

func fillString(returnString string, toLength int) string {
	for {
		lengtString := len(returnString)
		if lengtString < toLength {
			returnString = returnString + ":"
			continue
		}
		break
	}
	return returnString
}

func (fileNetwork *fileNetwork) RequestFile(filename string, address string) (string, int64, []byte) {
	connection, err := net.Dial("tcp", address)
	if err != nil {
		panic(err)
	}
	defer connection.Close()
	fmt.Println("sending fileID")
	connection.Write([]byte(filename))
	fmt.Println("Connected to server, start receiving the file name and file size")
	bufferFileName := make([]byte, 64)
	bufferFileSize := make([]byte, 10)

	connection.Read(bufferFileSize)
	fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

	connection.Read(bufferFileName)
	fileName := strings.Trim(string(bufferFileName), ":")

	var buf bytes.Buffer
	io.Copy(&buf, connection)
	return fileName, fileSize, buf.Bytes()
}
