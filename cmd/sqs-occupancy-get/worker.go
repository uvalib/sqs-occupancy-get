package main

import (
	"github.com/uvalib/virgo4-sqs-sdk/awssqs"
	"log"
	"net/http"
	"time"
)

// number of times to retry a message put before giving up and terminating
var sendRetries = uint(3)

func worker(workerId int, cameraClass string, client *http.Client, url string, pollTime int, aws awssqs.AWS_SQS, outQueue awssqs.QueueHandle) {

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
			} else {
				log.Printf("ERROR: worker %d received bad telemetry data, ignoring (%s)", workerId, payload)
			}

			// reset the block
			messages = messages[:0]
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

	opStatus1, err := aws.BatchMessagePut(outQueue, batch)
	if err != nil {
		// if an error we can handle, retry
		if err == awssqs.ErrOneOrMoreOperationsUnsuccessful {
			log.Printf("WARNING: worker %d one or more items failed to send to output queue, retrying...", workerId)

			// retry the failed items and bail out if we cannot retry
			err = aws.MessagePutRetry(outQueue, batch, opStatus1, sendRetries)
		}

		// bail out if an error and let someone else handle it
		if err != nil {
			return err
		}
	}

	return nil
}

//
// end of file
//
