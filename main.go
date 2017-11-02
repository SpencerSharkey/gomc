package main

import (
	"log"

	"github.com/spencersharkey/gomc/query"
)

// simple query test
func main() {
	req := query.NewRequest()

	err := req.Connect("mc.revive.gg:25565")
	if err != nil {
		log.Fatalln(err)
	}

	res, err := req.Simple()
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(res)
}
