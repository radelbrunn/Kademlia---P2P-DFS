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
	binAddr := ""
	hexXAddr := []rune(hexAddr[:])

	for i := 0; i < len(hexXAddr); i++ {
		switch string(hexXAddr[i]) {
		case "0":
			binAddr += "0000"
			break
		case "1":
			binAddr += "0001"
			break
		case "2":
			binAddr += "0010"
			break
		case "3":
			binAddr += "0011"
			break
		case "4":
			binAddr += "0100"
			break
		case "5":
			binAddr += "0101"
			break
		case "6":
			binAddr += "0110"
			break
		case "7":
			binAddr += "0111"
			break
		case "8":
			binAddr += "1000"
			break
		case "9":
			binAddr += "1001"
			break
		case "a":
			binAddr += "1010"
			break
		case "b":
			binAddr += "1011"
			break
		case "c":
			binAddr += "1100"
			break
		case "d":
			binAddr += "1101"
			break
		case "e":
			binAddr += "1110"
			break
		case "f":
			binAddr += "1111"
			break
		}
	}

	return binAddr
}
