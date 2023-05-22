package main

import (
	"encoding/json"
	"log"
	"strings"
)

// the camera details we need to identify if the telementry is suspect
type CameraTelemetry struct {
	Name      string `json:"name"`
	Serial    string `json:"serial"`
	Occupancy int    `json:"occupancy"`
	TotalIn   int    `json:"total_in"`
	TotalOut  int    `json:"total_out"`
}

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

func logSuspectTelemetry(workerId int, payload []byte) error {

	// extract the camera telemetry and determine if it is suspect
	var ct CameraTelemetry
	err := json.Unmarshal(payload, &ct)
	if err != nil {
		// nonsense data (cannot decode), just return the error
		return err
	}

	// potentially suspect telemetry
	if ct.Occupancy == -1 || ct.TotalIn == -1 || ct.TotalOut == -1 {
		log.Printf("WARNING: worker %d received suspect telemetry data from [%s/%s] (%s)", workerId, ct.Name, ct.Serial, payload)
	}

	return nil
}

//
// end of file
//
