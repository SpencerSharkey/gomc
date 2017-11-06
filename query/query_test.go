package query

import (
	"bytes"
	"log"
	"net"
	"reflect"
	"testing"
)

func checkFatalErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestSimpleQuery(t *testing.T) {

	go startQueryServer()

	req := NewRequest()

	err := req.Connect(":24565")
	checkFatalErr(t, err)

	res, err := req.Simple()
	checkFatalErr(t, err)

	validResponse := &SimpleResponse{
		MOTD:       "A Minecraft Server",
		GameType:   "SMP",
		Map:        "world",
		NumPlayers: 2,
		MaxPlayers: 20,
		HostPort:   25565,
		HostIP:     "127.0.0.1",
	}

	if !reflect.DeepEqual(validResponse, res) {
		t.Fatal("Simple query response invalid!", "\nexpected:\n\t", validResponse, "\nresult:\n\t", res)
	}
}

func fatalServerError(err error) {
	if err != nil {
		log.Fatalln("Pseudo server error!" + err.Error())
	}
}

func startQueryServer() {
	// Simple MC query emulator (server)
	go func() {
		addr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:24565")
		fatalServerError(err)

		serv, err := net.ListenUDP("udp", addr)
		fatalServerError(err)

		defer serv.Close()

		for {
			buf := make([]byte, 16)
			n, clientAddr, err := serv.ReadFromUDP(buf)
			msg := buf[:n]

			// Validate magic bytes
			if bytes.Compare(msg[:2], []byte{0xFE, 0xFD}) == 0 {
				msgType := msg[2]
				msgSessionID := msg[2:6]
				if msgType == 0x09 {
					// challenge token response
					payload := &bytes.Buffer{}
					payload.WriteByte(0x09)
					payload.Write(msgSessionID)
					payload.WriteString("1234")
					payload.WriteByte(0x00)
					_, err = serv.WriteToUDP(payload.Bytes(), clientAddr)
					fatalServerError(err)
				} else if msgType == 0x00 {
					payload := &bytes.Buffer{}
					payload.WriteByte(0x00)
					payload.Write(msgSessionID)
					parts := []string{
						"A Minecraft Server",
						"SMP",
						"world",
						"2",
						"20",
						"\xDD\x63127.0.0.1",
					}
					for _, part := range parts {
						payload.WriteString(part)
						payload.WriteByte(0x00)
					}
					_, err = serv.WriteToUDP(payload.Bytes(), clientAddr)
					fatalServerError(err)
				}
			}
			fatalServerError(err)
		}
	}()
}
