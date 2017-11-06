package main

import (
	"log"
	"os"

	"github.com/SpencerSharkey/gomc/query"
)

// simple query test
func main() {
	req := query.NewRequest()

	err := req.Connect(os.Args[1])
	if err != nil {
		log.Fatalln(err)
	}

	res, err := req.Simple()
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("%#v\n", res)
}
