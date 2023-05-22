package main

import (
	"log"
	"os"
	"strings"
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

	// load our AWS_SQS helper object
	aws, err := awssqs.NewAwsSqs(awssqs.AwsSqsConfig{MessageBucketName: cfg.MessageBucketName})
	fatalIfError(err)

	outQueueHandle, err := aws.QueueHandle(cfg.OutQueueName)
	fatalIfError(err)

	// start main camera workers here
	for ix := 0; ix < len(cfg.OccupancyIp); ix++ {
		client := newDigestClient(cfg.OccupancyUsername[ix], cfg.OccupancyPassword[ix], cfg.EndpointTimeout)
		url := strings.Replace(cfg.OccupancyQuery, "{:ip:}", cfg.OccupancyIp[ix], 1)
		go worker(ix, OccupancyCamera, client, url, cfg.PollTimeSeconds, aws, outQueueHandle)
	}

	// start all camera workers here
	for ix := 0; ix < len(cfg.SumIp); ix++ {
		client := newDigestClient(cfg.SumUsername[ix], cfg.SumPassword[ix], cfg.EndpointTimeout)
		url := strings.Replace(cfg.SumQuery, "{:ip:}", cfg.SumIp[ix], 1)
		go worker(ix+len(cfg.OccupancyIp), SumCamera, client, url, cfg.PollTimeSeconds, aws, outQueueHandle)
	}

	// sleep forever
	for {
		time.Sleep(999 * time.Second)
	}

	// should never get here
}

//
// end of file
//
