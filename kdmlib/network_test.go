package kdmlib

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestSendPacket(t *testing.T) {
	pc, err := net.ListenPacket("udp", ":8080")
	if err != nil {
		fmt.Println("ERROR")
	} else {
		buf := make([]byte, 1024)
		go func() {
			time.Sleep(time.Second)
			sendPacket("127.0.0.1", "8080", []byte("toto"))
		}()
		n, _, _ := pc.ReadFrom(buf)
		if string(buf[:n]) != "toto" {
			t.Error("wrong packet arrived")
		}

	}
}
