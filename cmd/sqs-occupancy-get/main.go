package main

import (
	"log"
	"os"
	"time"

	"github.com/uvalib/virgo4-sqs-sdk/awssqs"
)

// main entry point
func main() {

	log.Printf("===> %s service staring up (version: %s) <===", os.Args[0], Version())

	// Get config params and use them to init service context. Any issues are fatal
	cfg := LoadConfiguration()

	// load our AWS_SQS helper object
	aws, err := awssqs.NewAwsSqs(awssqs.AwsSqsConfig{MessageBucketName: cfg.MessageBucketName})
	fatalIfError(err)

	outQueueHandle, err := aws.QueueHandle(cfg.OutQueueName)
	fatalIfError(err)

	// start workers here
	for ix := 0; ix < len(cfg.Endpoints); ix++ {
		go worker(ix, cfg, aws, outQueueHandle)
	}

	// sleep forever
	for {
		time.Sleep(999 * time.Second)
	}
}

//
// end of file
//
