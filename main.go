package main


import (
	"encoding/json"
	"log"
	"os"

	"github.com/SpencerSharkey/gomc/query"
)

// Version - app release version
const Version string = "0.0.1"

func doSimpleQuery(addr string) {
	req := query.NewRequest()

	err := req.Connect(addr)
	if err != nil {
		log.Fatalln(err)
	}

	res, err := req.Simple()
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("=== SIMPLE QUERY ===")
	o, _ := json.MarshalIndent(res, "", "  ")
	log.Println(string(o))
}

func doFullQuery(addr string) {
	req := query.NewRequest()

	err := req.Connect(addr)
	if err != nil {
		log.Fatalln(err)
	}

	res, err := req.Full()
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("==== FULL QUERY ====")
	o, _ := json.MarshalIndent(res, "", "  ")
	log.Println(string(o))
}

// simple query test
func main() {
	doSimpleQuery(os.Args[1])
	doFullQuery(os.Args[1])
}
