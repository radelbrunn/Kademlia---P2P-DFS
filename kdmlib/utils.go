package kdmlib

import (
	"math"
	"math/rand"
	"strconv"
	"time"
)

const (
	K        = 20
	ALPHA    = 3
	IDLENGTH = 160
)

func GenerateRandID() string {

	id := ""
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < IDLENGTH; i++ {
		id += strconv.Itoa(rand.Intn(2))
	}

	return id
}

func ConvertToHexAddr(binAddr string) string {
	hexAddr := ""
	for i := 0; i < IDLENGTH/4; i++ {
		newPart := []rune(binAddr[(i * 4) : (i*4)+4])
		newIntPart := 0
		for j := 0; j < 4; j++ {
			if string(newPart[j]) == "1" {
				newIntPart += int(math.Pow(2, float64(3-j)))
			}
		}
		hexAddr += strconv.FormatInt(int64(newIntPart), 16)
	}
	return hexAddr
}

func GenerateIDFromHex(hexAddr string) string {
	//TODO implement

	hexAddr = ""
	return "1100101101101011001010111010010001100101000101111111001111100000010000010010000011000101011010010010101001011010000100101110100100001010001000101110110011010100"
}
