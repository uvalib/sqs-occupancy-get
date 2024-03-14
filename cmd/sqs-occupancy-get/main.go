package main

import (
	"log"
	"os"
	"time"

	"github.com/uvalib/virgo4-sqs-sdk/awssqs"
)

const (
	OccupancyCamera string = "occupancy"
	SumCamera              = "sum"
)

// main entry point
func main() {

	log.Printf("===> %s service staring up (version: %s) <===", os.Args[0], Version())

	// Get config params and use them to init service context. Any issues are fatal
	cfg := LoadConfiguration()

	log.Printf("[main] initializing SQS...")
	// load our AWS_SQS helper object
	aws, err := awssqs.NewAwsSqs(awssqs.AwsSqsConfig{MessageBucketName: cfg.MessageBucketName})
	fatalIfError(err)

	log.Printf("[main] getting queue handle...")
	outQueueHandle, err := aws.QueueHandle(cfg.OutQueueName)
	fatalIfError(err)

	// create counter object
	counter := Counter{}

	log.Printf("[main] starting workers...")
	// start main camera workers here
	for ix := 0; ix < len(cfg.OccupancyEndpoint); ix++ {
		client := newDigestClient(cfg.OccupancyUsername[ix], cfg.OccupancyPassword[ix], cfg.EndpointTimeout)
		go worker(ix, OccupancyCamera, client, cfg.OccupancyEndpoint[ix], cfg.PollTimeSeconds, aws, outQueueHandle, &counter)
	}

	// start all camera workers here
	for ix := 0; ix < len(cfg.SumEndpoint); ix++ {
		client := newDigestClient(cfg.SumUsername[ix], cfg.SumPassword[ix], cfg.EndpointTimeout)
		go worker(ix+len(cfg.OccupancyEndpoint), SumCamera, client, cfg.SumEndpoint[ix], cfg.PollTimeSeconds, aws, outQueueHandle, &counter)
	}

	// sleep and show metrics forever
	for {
		time.Sleep(time.Duration(cfg.PollTimeSeconds) * time.Second)
		s, e := counter.Get()
		log.Printf("[main] since startup, processed %d messages (success: %d, error: %d)",
			s+e, s, e)
	}

	// should never get here
}

//
// end of file
//
