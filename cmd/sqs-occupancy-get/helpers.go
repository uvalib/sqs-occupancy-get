package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

var location, _ = time.LoadLocation("America/New_York")
var suspectTelemetryError = fmt.Errorf("suspect telemetry data, one or more count fields is -1")

// the camera details we need to identify if the telementry is suspect
type CameraTelemetry struct {
	Timestamp string `json:"timestamp"`
	Occupancy int    `json:"occupancy,omitempty"` // some endpoints do not deliver this
	In        int    `json:"in"`
	Out       int    `json:"out"`
}

func fatalIfError(err error) {
	if err != nil {
		log.Fatalf("FATAL ERROR: %s", err.Error())
	}
}

func cleanupAndValidateJson(workerId int, cameraClass string, url string, payload []byte) ([]byte, error) {

	// special cases to make the data cleaner/legal
	pl := string(payload)
	pl = strings.Replace(pl, "\n", "", -1)
	pl = strings.Replace(pl, "  ", " ", -1)
	pl = strings.Replace(pl, "average visit time", "average_visit_time", 1)
	pl = strings.Replace(pl, "total in", "in", 1)
	pl = strings.Replace(pl, "total out", "out", 1)

	// convert to JSON structure to validate
	var ct CameraTelemetry
	err := json.Unmarshal(payload, &ct)
	if err != nil {
		// nonsense data (cannot decode), just return the error
		return payload, err
	}

	// add the camera class field
	s := fmt.Sprintf(", \"class\" : \"%s\"}", cameraClass)
	pl = strings.Replace(pl, "}", s, 1)

	// if the payload does not contain the occupancy (some payloads don't), we need to add it
	if strings.Contains(pl, "occupancy") == false {
		s := fmt.Sprintf(", \"occupancy\" : 0}")
		pl = strings.Replace(pl, "}", s, 1)
	}

	// if the payload does not contain the unixtime (some payloads don't), we need to add it
	if strings.Contains(pl, "unixtime") == false {
		format := "20060102150405" // yeah, crap right
		dt, err := time.ParseInLocation(format, ct.Timestamp, location)
		if err != nil {
			// nonsense date (cannot decode), just return the error
			return payload, err
		}
		s := fmt.Sprintf(", \"unixtime\" : %d}", dt.Unix())
		pl = strings.Replace(pl, "}", s, 1)
	}

	// potentially suspect telemetry from a main camera
	if cameraClass == OccupancyCamera && (ct.Occupancy == -1 || ct.In == -1 || ct.Out == -1) {
		//log.Printf("WARNING: [worker %d] received suspect telemetry data %s [%s]", workerId, url, payload)
		return payload, suspectTelemetryError
	}

	// potentially suspect telemetry from a normal camera
	if cameraClass == SumCamera && (ct.In == -1 || ct.Out == -1) {
		//log.Printf("WARNING: [worker %d] received suspect telemetry data %s [%s]", workerId, url, payload)
		return payload, suspectTelemetryError
	}

	return []byte(pl), nil
}

//
// end of file
//
