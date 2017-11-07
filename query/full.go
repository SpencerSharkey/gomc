/*
	Full implementation of Minecraft query-protocol client
	protocol spec: http://wiki.vg/Query
*/

package query

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
