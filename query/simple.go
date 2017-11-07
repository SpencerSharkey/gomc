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
	"strconv"
	"time"
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

	reader := bufio.NewReader(req.con)

	req.con.SetReadDeadline(time.Now().Add(req.readTimeout * time.Millisecond))
	reader.Discard(5) // Discard header data

	if reader.Buffered() < 5 {
		return nil, errors.New("malformed query response")
	}

	scan := bufio.NewScanner(reader)
	scan.Split(scanDelimittedResponse)

	scan.Scan()
	response.MOTD = scan.Text()

	scan.Scan()
	response.GameType = scan.Text()

	scan.Scan()
	response.Map = scan.Text()

	scan.Scan()
	response.NumPlayers, _ = strconv.Atoi(scan.Text())

	scan.Scan()
	response.MaxPlayers, _ = strconv.Atoi(scan.Text())

	scan.Scan()
	portAndIP := scan.Bytes()
	response.HostPort = int16(binary.LittleEndian.Uint16(portAndIP[:2]))
	response.HostIP = string(portAndIP[2:])
	req.con.SetReadDeadline(time.Time{})

	return response, nil
}
