/*
	Full implementation of Minecraft query-protocol client
	protocol spec: http://wiki.vg/Query
*/

package query

import (
	"io"
	"io/ioutil"
)

// FullResponse - Full Minecraft server query response
type FullResponse struct {
	Info    map[string]string `json:"info"`
	Players []string          `json:"players"`
}

// Full - Make a full query request
func (req *Request) Full() (*FullResponse, error) {
	response := &FullResponse{}

	challengeToken, err := req.GetChallengeToken()
	if err != nil {
		return nil, err
	}

	// Pad the request with 4 empty bytes to signify 'full' query request
	reqBuf := [15]byte{0xFE, 0xFD}
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

	// skip first 11 bytes of junk
	io.CopyN(ioutil.Discard, resBuf, 11)

	// parse key-value data into Info map
	response.Info = make(map[string]string)
	for {
		key, err := resBuf.ReadString(0x00)
		if err != nil {
			return response, err
		}

		if len(key) == 1 {
			break
		}

		value, err := resBuf.ReadString(0x00)
		if err != nil {
			return response, err
		}
		response.Info[key[:len(key)-1]] = value[:len(value)-1]
	}

	io.CopyN(ioutil.Discard, resBuf, 11)

	for {
		playerName, err := resBuf.ReadString(0x00)
		if err != nil {
			return response, err
		}
		if len(playerName) == 1 {
			break
		}
		response.Players = append(response.Players, playerName[:len(playerName)-1])
	}

	return response, nil

}
