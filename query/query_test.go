package query

import (
	"bytes"
	"log"
	"net"
	"reflect"
	"testing"
)

var (
	defaultQueryServerOptions QueryServerOptions
)

func checkFatalErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestSimpleQuery(t *testing.T) {
	serverOpen := make(chan string)
	go startQueryServer(serverOpen, &QueryServerOptions{
		negativeChallengeResponse: true,
	})
	serverAddr := <-serverOpen
	req := NewRequest()
	err := req.Connect(serverAddr)
	checkFatalErr(t, err)
	req.SetReadTimeout(20)
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
		t.Fatal("simple query response invalid", "\nexpected:\n\t", validResponse, "\nresult:\n\t", res)
	}
}

// Tests malformed challenge response error handling
func TestSimpleMalformedHeader(t *testing.T) {
	serverOpen := make(chan string)
	go startQueryServer(serverOpen, &QueryServerOptions{
		malformedResponseHeader: true,
	})
	serverAddr := <-serverOpen

	req := NewRequest()
	err := req.Connect(serverAddr)
	req.SetReadTimeout(20)

	checkFatalErr(t, err)

	_, err = req.Simple()
	if err == nil {
		t.Fatal("invalid challenge response should throw error")
	}
}

// Tests malformed challenge response error handling
func TestSimpleMalformedChallenge(t *testing.T) {
	serverOpen := make(chan string)
	go startQueryServer(serverOpen, &QueryServerOptions{
		malformedChallengeReponse: true,
	})
	serverAddr := <-serverOpen

	req := NewRequest()
	err := req.Connect(serverAddr)
	req.SetReadTimeout(20)

	checkFatalErr(t, err)

	_, err = req.Simple()
	if err == nil {
		t.Fatal("invalid challenge response should throw error")
	}
}

// Tests malformed simple query response error handling
func TestSimpleMalformedQuery(t *testing.T) {
	serverOpen := make(chan string)
	go startQueryServer(serverOpen, &QueryServerOptions{
		malformedQueryResponse: true,
	})
	serverAddr := <-serverOpen

	req := NewRequest()
	err := req.Connect(serverAddr)
	req.SetReadTimeout(20)

	checkFatalErr(t, err)

	_, err = req.Simple()
	if err == nil {
		t.Fatal("invalid query response should throw error")
	}
}

// Tests if errors are thrown when req.Connect isn't called
func TestSimpleQueryNoConnection(t *testing.T) {
	req := NewRequest()
	_, err := req.Simple()
	if err == nil {
		t.Fatal("invalid query response should throw error")
	}
}

// Tests timeout handling
func TestSimpleQueryTimeouts(t *testing.T) {
	// Challenge request timeout test
	serverOpen := make(chan string)
	go startQueryServer(serverOpen, &QueryServerOptions{
		simulateChallengeTimeout: true,
	})
	serverAddr := <-serverOpen

	req := NewRequest()
	req.Connect(serverAddr)
	req.SetReadTimeout(20)

	var err error
	_, err = req.Simple()
	if err == nil {
		t.Fatal("simple query challenge timeout test failed to produce error")
	}

	// Query request timeout test
	serverOpen = make(chan string)
	go startQueryServer(serverOpen, &QueryServerOptions{
		simulateQueryTimeout: true,
	})
	serverAddr = <-serverOpen

	req.Connect(serverAddr)
	req.SetReadTimeout(20)

	_, err = req.Simple()
	if err == nil {
		t.Fatal("simple query timeout test failed to produce error (malformed query response)")
	}

}

// Tests if we fail to resolve a hostname addr
func TestSimpleQueryResolveFail(t *testing.T) {
	req := NewRequest()
	err := req.Connect("foobar.foo:15976")
	if err == nil {
		t.Fatal("expecting req.Connect(<invalid hostname>) to throw error")
	}
}

/*
	startQueryServer(chan int)
		Starts a fake server able to respond to
		Minecraft query packets. Useful for testing!

		A better solution would be to include a
		more extensive query server in the library
		and test against that as well.
*/

type QueryServerOptions struct {
	malformedChallengeReponse bool
	negativeChallengeResponse bool
	malformedQueryResponse    bool
	malformedResponseHeader   bool
	simulateChallengeTimeout  bool
	simulateQueryTimeout      bool
}

func fatalServerError(err error) {
	if err != nil {
		log.Fatalln("Pseudo query server error! " + err.Error())
	}
}

func startQueryServer(serverOpen chan string, opts *QueryServerOptions) {
	addr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:0")
	fatalServerError(err)

	serv, err := net.ListenUDP("udp", addr)
	fatalServerError(err)
	defer serv.Close()

	// Send our assigned address to our caller
	// they will also know we are now accepting connections
	serverOpen <- serv.LocalAddr().String()

	// Read loop
	buf := make([]byte, 16)
	for {
		n, clientAddr, err := serv.ReadFromUDP(buf)
		msg := buf[:n]

		// Validate magic bytes
		if bytes.Compare(msg[:2], []byte{0xFE, 0xFD}) == 0 {
			msgType := msg[2]
			msgSessionID := msg[3:7]

			if opts.malformedResponseHeader == true {
				payload := &bytes.Buffer{}
				payload.WriteByte(0x09)
				payload.Write([]byte{0x00, 0x00, 0x00, 0x00})
				_, err = serv.WriteToUDP(payload.Bytes(), clientAddr)
				fatalServerError(err)
			}

			if msgType == 0x09 { // challenge token response

				if opts.simulateChallengeTimeout == true {
					continue
				}

				payload := &bytes.Buffer{}
				payload.WriteByte(0x09)
				payload.Write(msgSessionID)

				if opts.malformedChallengeReponse == true {
					payload.WriteByte(0x99)
				} else {
					if opts.negativeChallengeResponse == true {
						payload.WriteString("-")
					}
					payload.WriteString("9513307")
					payload.WriteByte(0x00)
				}

				_, err = serv.WriteToUDP(payload.Bytes(), clientAddr)
				fatalServerError(err)

			} else if msgType == 0x00 && n == 11 { // "simple" query identifier

				// Handle test cases (if set)
				if opts.malformedQueryResponse == true {
					// send some garbage
					serv.WriteToUDP([]byte{53, 156, 15, 123, 158}, clientAddr)
					continue
				}

				if opts.simulateQueryTimeout == true {
					continue
				}

				// Construct our valid response
				payload := &bytes.Buffer{}
				payload.WriteByte(0x00)
				payload.Write(msgSessionID)

				// Pack our fake server data
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

				// Send data
				_, err = serv.WriteToUDP(payload.Bytes(), clientAddr)
				fatalServerError(err)

			}
		}
		fatalServerError(err)
	}
}
