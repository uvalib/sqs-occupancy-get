package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

var maxCameras = 256

// ServiceConfig defines the service configuration parameters
type ServiceConfig struct {
	OutQueueName      string // SQS queue name for outbound documents
	MessageBucketName string // the bucket to use for large messages

	OccupancyEndpoint []string // the list of camera endpoints
	OccupancyUsername []string // the list of camera usernames
	OccupancyPassword []string // the list of camera passwords
	SumEndpoint       []string // the list of camera endpoints
	SumUsername       []string // the list of camera usernames
	SumPassword       []string // the list of camera passwords

	EndpointTimeout int // the endpoint timeout in seconds
	PollTimeSeconds int // the endpoint poll time in seconds
}

func ensureSet(env string) string {
	val, set := os.LookupEnv(env)

	if set == false {
		log.Printf("FATAL ERROR: environment variable not set: [%s]", env)
		os.Exit(1)
	}

	return val
}

func ensureSetAndNonEmpty(env string) string {
	val := ensureSet(env)

	if val == "" {
		log.Printf("FATAL ERROR: environment variable not set: [%s]", env)
		os.Exit(1)
	}

	return val
}

func envToInt(env string) int {

	number := ensureSetAndNonEmpty(env)
	n, err := strconv.Atoi(number)
	fatalIfError(err)
	return n
}

// LoadConfiguration will load the service configuration from env/cmdline
// and return a pointer to it. Any failures are fatal.
func LoadConfiguration() *ServiceConfig {

	var cfg ServiceConfig

	cfg.OutQueueName = ensureSetAndNonEmpty("SQS_OCCUPANCY_GET_OUT_QUEUE")
	cfg.MessageBucketName = ensureSetAndNonEmpty("SQS_MESSAGE_BUCKET")
	cfg.EndpointTimeout = envToInt("SQS_OCCUPANCY_GET_CAMERA_TIMEOUT")
	cfg.PollTimeSeconds = envToInt("SQS_OCCUPANCY_GET_POLL_IN_SECONDS")

	userName := ensureSetAndNonEmpty("SQS_OCCUPANCY_GET_USERNAME")
	password := ensureSetAndNonEmpty("SQS_OCCUPANCY_GET_PASSWORD")

	for ix := 0; ix < maxCameras; ix++ {
		env := fmt.Sprintf("SQS_OCCUPANCY_GET_MAIN_CAMERA_%03d", ix+1)
		val, set := os.LookupEnv(env)
		if set == true {
			cfg.OccupancyEndpoint = append(cfg.OccupancyEndpoint, val)
			cfg.OccupancyUsername = append(cfg.OccupancyUsername, userName)
			cfg.OccupancyPassword = append(cfg.OccupancyPassword, password)
		} else {
			break
		}
	}

	for ix := 0; ix < maxCameras; ix++ {
		env := fmt.Sprintf("SQS_OCCUPANCY_GET_CAMERA_%03d", ix+1)
		val, set := os.LookupEnv(env)
		if set == true {
			cfg.SumEndpoint = append(cfg.SumEndpoint, val)
			cfg.SumUsername = append(cfg.SumUsername, userName)
			cfg.SumPassword = append(cfg.SumPassword, password)
		} else {
			break
		}
	}

	// ensure we have 1 or more endpoints defined
	if len(cfg.OccupancyEndpoint) == 0 && len(cfg.SumEndpoint) == 0 {
		fatalIfError(fmt.Errorf("no camera ip addresses defined"))
	}

	log.Printf("[config] OutQueueName         = [%s]", cfg.OutQueueName)
	log.Printf("[config] MessageBucketName    = [%s]", cfg.MessageBucketName)
	log.Printf("[config] EndpointTimeout      = [%d]", cfg.EndpointTimeout)
	log.Printf("[config] PollTimeSeconds      = [%d]", cfg.PollTimeSeconds)

	for ix, _ := range cfg.OccupancyEndpoint {
		log.Printf("[config] Main camera %03d      = [%s (%s/REDACTED)]", ix+1, cfg.OccupancyEndpoint[ix], cfg.OccupancyUsername[ix])
	}

	for ix, _ := range cfg.SumEndpoint {
		log.Printf("[config] Camera %03d           = [%s (%s/REDACTED)]", ix+1, cfg.SumEndpoint[ix], cfg.SumUsername[ix])
	}

	return &cfg
}

//
// end of file
//
