/*
	Full implementation of Minecraft query-protocol client
	protocol spec: http://wiki.vg/Query
*/

package query

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math/rand"
	"net"
	"strconv"
	"time"
)

var (
	magicHeader = &[]byte{0xFE, 0xFD}
)

// SimpleResponse - Simple Minecraft server query response
type SimpleResponse struct {
	MOTD       string
	GameType   string
	Map        string
	NumPlayers int
	MaxPlayers int
	HostPort   int16
	HostIP     string
}

// FullResponse - Full Minecraft server query response
type FullResponse struct {
	MOTD       string
	GameType   string
	GameID     string
	Version    string
	Plugins    string
	Map        string
	NumPlayers int
	MaxPlayers int
	HostPort   int16
	HostIP     string
	Players    []string
}

// Request - Query Client
type Request struct {
	con         *net.UDPConn
	readTimeout time.Duration
	sesssionID  int32
}

// NewRequest - Query request factory
func NewRequest() *Request {
	req := &Request{}
	return req
}

// Connect initiates a new connection to the Minecraft server host
func (req *Request) Connect(hostaddr string) error {
	addr, err := net.ResolveUDPAddr("udp4", hostaddr)
	if err != nil {
		return errors.New("error resolving host")
	}

	req.con, err = net.DialUDP("udp4", nil, addr)
	if err != nil {
		return errors.New("error dialing UDP connection")
	}

	// set default read timeout
	req.readTimeout = 5000

	// Generate a Session ID
	req.sesssionID = GenerateSessionID()

	return nil
}

// GetChallengeToken - Retrives a challenge token from the server
func (req *Request) GetChallengeToken() (int32, error) {
	if req.con == nil {
		return -1, errors.New("no connection, call Request.Connect first")
	}

	// Build challenge token request packet
	buf := &bytes.Buffer{}
	buf.Write(*magicHeader)
	buf.WriteByte(0x09) // Packet Type 0x09 = Challenge Request
	binary.Write(buf, binary.BigEndian, req.sesssionID)

	req.con.Write(buf.Bytes())

	res, err := req.ReadWithDeadline(24, req.readTimeout)
	if err != nil {
		return -1, err
	}

	// Parse challenge response
	challengeToken, err := strconv.ParseUint(string(res[5:bytes.IndexByte(res[5:], 0x00)+5]), 10, 32)
	if err != nil {
		return -1, errors.New("error parsing challenge response from server")
	}

	return int32(challengeToken), nil
}

// ReadWithDeadline will read from our socket with a specified timeout
func (req *Request) ReadWithDeadline(length int, timeout time.Duration) ([]byte, error) {
	res := make([]byte, length)
	req.con.SetDeadline(time.Now().Add(timeout * time.Millisecond))
	bytes, err := req.con.Read(res)
	req.con.SetDeadline(time.Time{})
	if bytes == 0 || err != nil {
		return nil, errors.New("timeout of " + strconv.Itoa(int(timeout)) + "ms exceeded when reading from server (" + err.Error() + ")")
	}
	return res, nil
}

// GenerateSessionID - Generates a 32-bit SessionID
func GenerateSessionID() int32 {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Int31() & 0x0F0F0F0F
}

// A simple scanner func to read our null-byte delimited data
func scanDelimittedResponse(input []byte, eof bool) (adv int, token []byte, err error) {
	if len(input) == 0 {
		return 0, nil, errors.New("end of input")
	}
	i := bytes.Index(input, []byte{0x00})
	return i + 1, input[:i], nil
}
