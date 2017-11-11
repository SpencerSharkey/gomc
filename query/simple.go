/*
	Full implementation of Minecraft query-protocol client
	protocol spec: http://wiki.vg/Query
*/

package query

import (
	"bufio"
	"encoding/binary"
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

	challengeToken, err := req.GetChallengeToken()
	if err != nil {
		return nil, err
	}

	reqBuf := [11]byte{0xFE, 0xFD}
	copy(reqBuf[3:], req.sessionID[0:])
	copy(reqBuf[7:], challengeToken)
	req.con.Write(reqBuf[:])

	resBuf, err := req.ReadWithDeadline()
	if err != nil {
		return response, err
	}

	err = req.VerifyResponseHeader(resBuf)
	if err != nil {
		return response, err
	}

	scan := bufio.NewScanner(resBuf)
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
