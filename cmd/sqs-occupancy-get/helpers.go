package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var location, _ = time.LoadLocation("America/New_York")
var suspectTelemetryError = fmt.Errorf("suspect telemetry data, one or more count fields is -1")

// LegacyInboundCameraTelemetry -- inbound camera telemetry from legacy software
type LegacyInboundCameraTelemetry struct {
	Name      string `json:"name"`
	Serial    string `json:"serial"`
	Unixtime  int64  `json:"unixtime"` // not all requests include this
	Timestamp string `json:"timestamp"`
	Occupancy int    `json:"occupancy"`
	In        int    `json:"in"`
	Out       int    `json:"out"`
}

// NewInboundCameraTelemetry -- inbound camera telemetry from the latest software
type NewInboundCameraTelemetry struct {
	Data []NewInboundCameraPayload `json:"data"`
}

// NewInboundCameraPayload --
type NewInboundCameraPayload struct {
	Start string `json:"start"`
	End   string `json:"end"`
	In    int    `json:"in"`
	Out   int    `json:"out"`
}

// OutboundCameraTelemetry -- the normalized outbound structure
type OutboundCameraTelemetry struct {
	Class     string `json:"class"`
	Serial    string `json:"serial"`
	Unixtime  int64  `json:"unixtime"`
	Occupancy int    `json:"occupancy"`
	In        int    `json:"in"`
	Out       int    `json:"out"`
}

func fatalIfError(err error) {
	if err != nil {
		log.Fatalf("FATAL ERROR: %s", err.Error())
	}
}

// inbound json may be illegal so attempt to clean up
func cleanupInboundJson(payload []byte) []byte {

	// special cases to make the legacy camera telemetry data legal JSON
	pl := string(payload)
	pl = strings.Replace(pl, "\n", "", -1)
	pl = strings.Replace(pl, "  ", " ", -1)
	pl = strings.Replace(pl, "average visit time", "average_visit_time", 1)
	pl = strings.Replace(pl, "total in", "in", 1)
	pl = strings.Replace(pl, "total out", "out", 1)

	// normalize a couple of responses
	pl = strings.Replace(pl, "total_in", "in", 1)
	pl = strings.Replace(pl, "total_out", "out", 1)

	return []byte(pl)
}

func makeOutboundTelemetry(workerId int, cameraClass string, payload []byte) (OutboundCameraTelemetry, error) {

	// outbound structure
	var outbound OutboundCameraTelemetry
	// add the camera class
	outbound.Class = cameraClass

	// first try the newer structure
	var current NewInboundCameraTelemetry
	err := json.Unmarshal(payload, &current)
	if err != nil {
		// nonsense data (cannot decode), just return the error
		return outbound, err
	}

	// looks like we have a winner
	if len(current.Data) != 0 {
		outbound.In = current.Data[0].In
		outbound.Out = current.Data[0].Out
		outbound.Unixtime = time.Now().Unix()

		// note this payload does NOT include a serial number... we have to handle this later

		// and we are done
		return outbound, nil
	}

	// now try legacy telemetry
	var legacy LegacyInboundCameraTelemetry
	err = json.Unmarshal(payload, &legacy)
	if err != nil {
		// nonsense data (cannot decode), just return the error
		return outbound, err
	}

	// if the payload does not contain the unixtime (some payloads don't), we need to add it
	if legacy.Unixtime == 0 {

		var err error
		dt := time.Now()

		// if we have a timestamp then use it otherwise use our local time
		if len(legacy.Timestamp) != 0 {

			// try multiple formats
			format := "20060102150405" // yeah, crap right
			dt, err = time.ParseInLocation(format, legacy.Timestamp, location)
			if err != nil {
				format := "2006-01-02T15:04:05+00:00" // yeah, crap right
				dt, err = time.ParseInLocation(format, legacy.Timestamp, location)
				if err != nil {
					// nonsense date (cannot decode), just return the error
					return outbound, err
				}
			}
		}
		legacy.Unixtime = dt.Unix()
	}

	// populate the outbound structure
	outbound.Serial = legacy.Serial
	outbound.Unixtime = legacy.Unixtime
	outbound.Occupancy = legacy.Occupancy
	outbound.In = legacy.In
	outbound.Out = legacy.Out

	// potentially suspect telemetry from a main camera
	if outbound.Class == OccupancyCamera && (outbound.Occupancy == -1 || outbound.In == -1 || outbound.Out == -1) {
		//log.Printf("WARNING: [worker %d] received suspect telemetry data %s [%s]", workerId, url, payload)
		return outbound, suspectTelemetryError
	}

	// potentially suspect telemetry from a normal camera
	if outbound.Class == SumCamera && (outbound.In == -1 || outbound.Out == -1) {
		//log.Printf("WARNING: [worker %d] received suspect telemetry data %s [%s]", workerId, url, payload)
		return outbound, suspectTelemetryError
	}

	// return the outbound
	return outbound, nil
}

func getSerial(workerId int, client *http.Client, url string) string {

	serial := ""
	// note, this is a special case...
	re := regexp.MustCompile("export.*$")
	newUrl := re.ReplaceAllString(url, "occupancy")

	payload, err := httpGet(workerId, newUrl, client)
	if err == nil {
		var legacy LegacyInboundCameraTelemetry
		err = json.Unmarshal(payload, &legacy)
		if err == nil {
			return legacy.Serial
		}
	}

	return serial
}

//
// end of file
//
