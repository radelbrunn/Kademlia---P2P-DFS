package kdmlib

import (
	"fmt"
	"testing"
)

func TestConvertToHexAddr(t *testing.T) {
	id := "1010101001011010111011000010101001110001100011001011111111010101010110100111111101011110111010010100110101111000000110100101111101111100101000000100111010011110"
	hexId := ConvertToHexAddr(id)

	if hexId != "aa5aec2a718cbfd55a7f5ee94d781a5f7ca04e9e" {
		t.Error("Expected aa5aec2a718cbfd55a7f5ee94d781a5f7ca04e9e, got ", hexId)
	}

	id = "0100110111001100111000100101110111110100111001010101010110000101100011110110111000111101001001100000000111111000001111010001011010000100100000000100001000101111"
	hexId = ConvertToHexAddr(id)

	if hexId != "4dcce25df4e555858f6e3d2601f83d168480422f" {
		t.Error("Expected 4dcce25df4e555858f6e3d2601f83d168480422f, got ", hexId)
	}

	if len(id)/len(hexId) != 4 {
		t.Error("Wrong ID lengths")
	}
}

func TestGenerateIDFromHex(t *testing.T) {
	hexId := "82aa9fddde0a066ecdb5dc67d27ef51e6afa6862"
	id := GenerateIDFromHex(hexId)

	if id != "1000001010101010100111111101110111011110000010100000011001101110110011011011010111011100011001111101001001111110111101010001111001101010111110100110100001100010" {
		t.Error("Wrong binary ID returned")
	}

	hexId = "67d46d010f4b047962d14f4d3a8cb4751fcc9d1c"
	id = GenerateIDFromHex(hexId)

	if id != "0110011111010100011011010000000100001111010010110000010001111001011000101101000101001111010011010011101010001100101101000111010100011111110011001001110100011100" {
		t.Error("Wrong binary ID returned")
	}

	if len(id)/len(hexId) != 4 {
		t.Error("Wrong ID lengths")
	}
}

func TestComputeDistance(t *testing.T) {
	res, _ := ComputeDistance("010", "111")
	if res != "101" {
		t.Error("not the right result")
	}

	_, err := ComputeDistance("11", "1111")
	if err == nil {
		t.Error("expected an error")
	}
}

func TestConvertToUDPAddr(t *testing.T) {
	t1 := AddressTriple{"111.1.1.1", "9000", "1001"}
	addr := ConvertToUDPAddr(t1)

	if addr.String() != t1.Ip+":"+t1.Port {
		t.Error("Expected", t1.Ip, ":", t1.Port, ", got ", addr.String())
	}
}

func TestGenerateZeroID(t *testing.T) {
	zeroID := GenerateZeroID(160)
	hexZero := ConvertToHexAddr(zeroID)

	if len(hexZero) != 40 {
		fmt.Println("Wrong ID length")
	}
}
