package main

import (
	"encoding/json"
	"fmt"
	"github.com/uvalib/virgo4-sqs-sdk/awssqs"
	"log"
	"net/http"
	"strings"
	"time"
)

// number of times to retry a message put before giving up and terminating
var sendRetries = 3
var retrySleep = 500 * time.Millisecond

func worker(workerId int, cameraClass string, client *http.Client, url string, pollTime int, aws awssqs.AWS_SQS, outQueue awssqs.QueueHandle, counter *Counter) {

	messages := make([]awssqs.Message, 0, 1)
	serial := "" // some responses don't contain the serial number and we need to save it between calls

	for {
		payload, err := httpGet(workerId, urlRewrite(url), client)
		if err == nil {
			//fmt.Printf("IN [%s]\n", string(payload))

			payload = cleanupInboundJson(payload)

			// normalize the payload string and log if it looks suspect
			obt, err := makeOutboundTelemetry(workerId, cameraClass, payload)
			if err == nil {

				// some responses do not include the serial number
				if len(obt.Serial) == 0 {
					if len(serial) == 0 {
						// this is a special case and I don't like it
						log.Printf("INFO: [worker %d] getting serial number...", workerId)
						serial = getSerial(workerId, client, url)
					}

					obt.Serial = serial
				}

				pl, err := json.Marshal(&obt)
				if err == nil {
					//fmt.Printf("OUT [%s]\n", string(pl))
					// send to the SQS queue
					messages = append(messages, constructMessage(pl))
					err = sendOutboundMessages(workerId, aws, outQueue, messages)
					fatalIfError(err)
					counter.AddSuccess(1)
				} else {
					log.Printf("ERROR: [worker %d] received suspect telemetry data (%s), ignoring [%s]", workerId, err.Error(), payload)
					counter.AddError(1)
				}
			} else {
				log.Printf("ERROR: [worker %d] received suspect telemetry data (%s), ignoring [%s]", workerId, err.Error(), payload)
				counter.AddError(1)
			}

			// reset the block
			messages = messages[:0]
		} else {
			counter.AddError(1)
		}

		// and sleep...
		time.Sleep(time.Duration(pollTime) * time.Second)
	}

	// should never get here
}

func constructMessage(payload []byte) awssqs.Message {

	attributes := make([]awssqs.Attribute, 0, 2)
	attributes = append(attributes, awssqs.Attribute{Name: awssqs.AttributeKeyRecordType, Value: "json"})
	attributes = append(attributes, awssqs.Attribute{Name: awssqs.AttributeKeyRecordOperation, Value: awssqs.AttributeValueRecordOperationUpdate})
	return awssqs.Message{Attribs: attributes, Payload: payload}
}

func sendOutboundMessages(workerId int, aws awssqs.AWS_SQS, outQueue awssqs.QueueHandle, batch []awssqs.Message) error {

	// we only ever send 1 message so can implement a retry loop here
	attempt := 0
	for {
		_, err := aws.BatchMessagePut(outQueue, batch)
		if err == nil {
			return nil
		}

		// is it time to give up
		attempt++
		if attempt == sendRetries {
			log.Printf("ERROR: [worker %d] failed to send %d times, giving up", workerId, sendRetries)
			return err
		}

		// wait a bit and try again
		log.Printf("ERROR: [worker %d] failed to send %d time(s), waiting to retry", workerId, attempt)
		time.Sleep(retrySleep)
	}
}

func urlRewrite(url string) string {

	res := url
	// do some URL substitution
	// look for tomorrow's date placeholder
	if strings.Contains(url, "XX_TOMORROW_YYMMDD_XX") {
		today := time.Now()
		tomorrow := today.AddDate(0, 0, 1)
		yymmdd := fmt.Sprintf("%04d%02d%02d", tomorrow.Year(), tomorrow.Month(), tomorrow.Day())
		res = strings.Replace(url, "XX_TOMORROW_YYMMDD_XX", yymmdd, 1)
	}

	return res
}

//
// end of file
//
