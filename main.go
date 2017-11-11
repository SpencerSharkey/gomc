package main

import (
	"log"
	"os"

	"github.com/SpencerSharkey/gomc/query"
)

// Version - app release version
const Version string = "0.0.1"

func doQuery() {
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

// simple query test
func main() {
	doQuery()
}
