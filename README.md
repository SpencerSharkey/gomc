Minecraft Go
===
[![Build Status](https://travis-ci.org/SpencerSharkey/gomc.svg?branch=master)](https://travis-ci.org/SpencerSharkey/gomc)
[![Coverage Status](https://coveralls.io/repos/github/SpencerSharkey/gomc/badge.svg?branch=master)](https://coveralls.io/github/SpencerSharkey/gomc?branch=master)
[![Maintainability](https://api.codeclimate.com/v1/badges/9b0cf6b594cbf294dd5c/maintainability)](https://codeclimate.com/github/SpencerSharkey/gomc/maintainability)
[![HitCount](http://hits.dwyl.io/SpencerSharkey/gomc.svg)](http://hits.dwyl.io/SpencerSharkey/gomc)

## Introduction
Go (golang) library for Minecraft utilities. 

The only functionality currently implemented is a query client compatible with the vanilla Minecraft [query protocol](http://wiki.vg/Query).

Planned packages:
* ~~query~~ [/query](https://github.com/SpencerSharkey/gomc/tree/master/query)
* rcon

**Note:** This package is being developed for use within the [spencersharkey/mcapi](https://github.com/spencersharkey/mcapi) application, aimed at exposing an HTTP client for various Minecraft-related utilities.

## "query" Package Documentation

### (*query.Request) Simple()
```golang
type SimpleResponse struct {
	Hostname   string `json:"hostname"`
	GameType   string `json:"gametype"`
	Map        string `json:"map"`
	NumPlayers int    `json:"numplayers"`
	MaxPlayers int    `json:"maxplayers"`
	HostPort   int16  `json:"hostport"`
	HostIP     string `json:"hostip"`
}
```
Simple request example:
```golang
req := query.NewRequest()
err := req.Connect("127.0.0.1:25565")
res, err := req.Simple()
```
```json
{
  "hostname": "Revive Minecraft! http://revive.gg/chat",
  "gametype": "SMP",
  "map": "world",
  "numplayers": 4,
  "maxplayers": 128,
  "hostport": 25565,
  "hostip": "127.0.0.1"
}
```
### (*query.Request) Full()
*Note:* `FullResponse` embeds the `SimpleResponse` struct
```golang
type FullResponse struct {
	SimpleResponse
	Info    map[string]string `json:"info"`
	Players []string          `json:"players"`
}
```
Full request example:
```golang
req := query.NewRequest()
err := req.Connect("127.0.0.1:25565")
res, err := req.Full()
```
```json
{
  "hostname": "Revive Minecraft! http://revive.gg/chat",
  "gametype": "SMP",
  "map": "world",
  "numplayers": 4,
  "maxplayers": 128,
  "hostport": 25565,
  "hostip": "127.0.0.1",
  "info": {
    "game_id": "MINECRAFT",
    "plugins": "",
    "version": "1.12.2"
  },
  "players": [
    "SpencerSharkey",
    "RumBull",
    "lonemajestyk",
    "nakkirakki"
  ]
}
```
