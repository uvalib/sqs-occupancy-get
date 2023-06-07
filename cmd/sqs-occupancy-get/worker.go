package main

import (
	"github.com/uvalib/virgo4-sqs-sdk/awssqs"
	"log"
	"net/http"
	"time"
)

// number of times to retry a message put before giving up and terminating
var sendRetries = 3
var retrySleep = 500 * time.Millisecond

func worker(workerId int, cameraClass string, client *http.Client, url string, pollTime int, aws awssqs.AWS_SQS, outQueue awssqs.QueueHandle, counter *Counter) {

	messages := make([]awssqs.Message, 0, 1)
	for {

		payload, err := httpGet(workerId, url, client)
		if err == nil {
			// cleanup the JSON string and log if it looks suspect
			payload, err = cleanupAndValidateJson(workerId, cameraClass, url, payload)
			if err == nil {
				// send to the SQS queue
				messages = append(messages, constructMessage(payload))
				err = sendOutboundMessages(workerId, aws, outQueue, messages)
				fatalIfError(err)
				counter.AddSuccess(1)
			} else {
				log.Printf("ERROR: [worker %d] received bad telemetry data (%s), ignoring [%s]", workerId, err.Error(), payload)
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

//
// end of file
//
