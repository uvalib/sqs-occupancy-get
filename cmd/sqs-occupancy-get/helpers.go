package main

import (
	"log"
	"strings"
)

func fatalIfError(err error) {
	if err != nil {
		log.Fatalf("FATAL ERROR: %s", err.Error())
	}
}

func convertLegalJson(payload []byte) []byte {

	// special cases to make the data more better
	pl := string(payload)
	pl = strings.Replace(pl, "\n", "", -1)
	pl = strings.Replace(pl, "  ", " ", -1)
	pl = strings.Replace(pl, "average visit time", "average_visit_time", 1)
	pl = strings.Replace(pl, "total in", "total_in", 1)
	pl = strings.Replace(pl, "total out", "total_out", 1)
	return []byte(pl)
}

//
// end of file
//
