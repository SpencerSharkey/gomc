/*
	Full implementation of Minecraft query-protocol client
	protocol spec: http://wiki.vg/Query
*/

package query

import (
	"bufio"
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

// ReadWithDeadline will read from our socket with a specified timeout
func (req *Request) ReadWithDeadline(length int, timeout time.Duration) ([]byte, error) {
	res := make([]byte, length)
	req.con.SetDeadline(time.Now().Add(timeout * time.Second))
	bytes, err := req.con.Read(res)
	req.con.SetDeadline(time.Time{})
	if bytes == 0 || err != nil {
		return nil, errors.New("timeout of " + strconv.Itoa(int(timeout)) + "s exceeded when reading from server")
	}
	return res, nil
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

// A simple scanner func to read our null-byte delimited data
func scanDelimittedResponse(input []byte, eof bool) (adv int, token []byte, err error) {
	i := bytes.Index(input, []byte{0x00})
	return i + 1, input[:i], nil
}

// Simple - Make a simple query request
func (req *Request) Simple() (*SimpleResponse, error) {
	response := &SimpleResponse{}

	if req.con == nil {
		return response, errors.New("no connection, call Request.Connect first")
	}

	challengeToken, err := req.GetChallengeToken()
	if err != nil {
		return nil, err
	}

	// Build simple query request packet
	buf := &bytes.Buffer{}
	buf.Write(*magicHeader)
	buf.WriteByte(0x00) // Packet Type 0x00 = Query Request
	binary.Write(buf, binary.BigEndian, req.sesssionID)
	binary.Write(buf, binary.BigEndian, challengeToken)

	req.con.Write(buf.Bytes())

	// Read and parse query data
	req.con.SetReadDeadline(time.Now().Add(req.readTimeout * time.Millisecond))

	reader := bufio.NewReader(req.con)
	reader.Discard(5) // Discard header data
	scan := bufio.NewScanner(reader)
	scan.Split(scanDelimittedResponse)

	scan.Scan()
	response.MOTD = scan.Text()

	scan.Scan()
	response.GameType = scan.Text()

	scan.Scan()
	response.Map = scan.Text()

	scan.Scan()
	response.NumPlayers, err = strconv.Atoi(scan.Text())
	if err != nil {
		return nil, errors.New("error parsing numplayers field")
	}

	scan.Scan()
	response.MaxPlayers, err = strconv.Atoi(scan.Text())
	if err != nil {
		return nil, errors.New("error parsing maxplayers field")
	}

	scan.Scan()
	portAndIP := scan.Bytes()
	response.HostPort = int16(binary.LittleEndian.Uint16(portAndIP[:2]))
	response.HostIP = string(portAndIP[2:])
	req.con.SetReadDeadline(time.Time{})

	return response, nil
}

// NewRequest - Query request factory
func NewRequest() *Request {
	req := &Request{}
	return req
}

// GenerateSessionID - Generates a 32-bit SessionID
func GenerateSessionID() int32 {
	rand.Seed(time.Now().UTC().UnixNano())
	return rand.Int31() & 0x0F0F0F0F
}
