// serverTCP project main.go
package main

import (
	//"bufio"
	//"encoding/hex"
	"fmt"
	"io"
	"net"
	//"strconv"
	"bytes"
	"encoding/binary"
	//"strings"
	//"time"
	"errors"

	"github.com/praktin/ModbusEmulatorServer/ModbusServer/CRC"
)

type modbus struct {
	Slaveid  byte
	Fnc      byte
	Startadr uint16
	Count    uint16
	Crc      [2]byte
}

type res struct {
	Id        byte
	Fnc       byte
	NumByte   uint16
	Nextbytes []byte
	Crc 	  [2]byte
}

func makeBytes(source uint16) []byte {
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], source)
	return b[:]
}

func getPhisicalAddress(logicalAddress uint16, function byte) (int, error) {
	switch function {
	case 1: //test  phisAddress>=10001
		return int(logicalAddress + 10001), nil
	case 3:
		return int(logicalAddress + 40001), nil
	default:
		return 0, errors.New("illegal function")
	}
}



func main() {

	x := make(map[int]int)
	x[40065] = 98
	x[40066] = 36
	x[40067] = 112
	x[40068] = 64
	x[40069] = 87
	x[40070] = 54
	x[40071] = 14
	x[40072] = 81
	x[40073] = 49

	fmt.Println("Launch server...")
	ln, err := net.Listen("tcp", ":8081")
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	conn, err := ln.Accept()
	if err != nil {
		fmt.Print("Client close")
	}

	var inputbytes [1024]byte

	for {
		q, err := io.ReadAtLeast(conn, inputbytes[:], 1)
		if err != nil {
			return
		}
		mes := inputbytes[:q]
		fmt.Printf("Message: %x", mes)

		p := modbus{}

		buf := bytes.NewReader(mes[:])

		err1 := binary.Read(buf, binary.BigEndian, &p)

		if err1 != nil {
			fmt.Println("Binary.Read failed: ", err1)
		}

		prov := mes[0:len(mes) - 2]
		var crcInput crc.Crc

		crcInput.Reset().PushBytes(prov)
		/*fmt.Printf("  prov: %x ", prov)
		fmt.Println("crcInput: ", crcInput)

		fmt.Println("p: ", p)*/

		checksum := uint16(mes[len(mes)-1])<<8 | uint16(mes[len(mes)-2])

		//fmt.Println("crcInput: ", crcInput.Value())
		//fmt.Println("mes[crc]: ", checksum)

		if checksum != crcInput.Value() {
			err =   errors.New("response CrcInput does not match")
			return
		}

		if p.Fnc != 3 {
			fmt.Println(p.Fnc)
			fmt.Println("Invalid function")
			return
		}

		m1 := res{}

		m1.Id = p.Slaveid
		m1.Fnc = p.Fnc
		m1.NumByte = p.Count * 2

		for i := int(p.Startadr); i < int(p.Startadr+p.Count); i++ {
			phisAddress, err := getPhisicalAddress(uint16(i), p.Fnc)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			if _, ok := x[phisAddress]; ok {
				s1 := makeBytes(uint16(x[phisAddress]))
				m1.Nextbytes = append(m1.Nextbytes, s1...)
			} else {
				fmt.Println("Invalid Address")
				return
			}
		}

		var result1 []byte
		var crcOutput crc.Crc
		crcMass:= make([]byte, 2)

		result1 = append(result1, m1.Id)
		result1 = append(result1, m1.Fnc)
		result1 = append(result1, byte(m1.NumByte))
		result1 = append(result1, m1.Nextbytes...)

		//fmt.Println("result: ", result1)

		crcOutput.Reset().PushBytes(result1)
		//fmt.Println("crc output: ", crcOutput)

		binary.LittleEndian.PutUint16(crcMass, crcOutput.Value())
		//fmt.Println("crcMass: ", crcMass)

		result1 = append(result1, crcMass[0])
		result1 = append(result1, crcMass[1])

		if err != nil {
			fmt.Println("Server close")
			return
		}

		fmt.Printf(" ### Message Received: %x", result1)
		//newmessege := bytes.ToUpper(result1)
		//conn.Write(newmessege + result1)
	}
}
