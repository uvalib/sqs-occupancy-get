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

	OccupancyQuery string // the get occupancy endpoint query
	SumQuery       string // the get sum in/out endpoint query

	OccupancyIp       []string // the list of camera ip addresses (occupancy queries)
	OccupancyUsername []string // the list of camera usernames
	OccupancyPassword []string // the list of camera passwords
	SumIp             []string // the list of camera ip addresses (sum queries)
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
	cfg.OccupancyQuery = ensureSetAndNonEmpty("SQS_OCCUPANCY_GET_OCCUPANCY_QUERY")
	cfg.SumQuery = ensureSetAndNonEmpty("SQS_OCCUPANCY_GET_SUM_QUERY")

	userName := ensureSetAndNonEmpty("SQS_OCCUPANCY_GET_USERNAME")
	password := ensureSetAndNonEmpty("SQS_OCCUPANCY_GET_PASSWORD")

	for ix := 0; ix < maxCameras; ix++ {
		env := fmt.Sprintf("SQS_OCCUPANCY_GET_MAIN_CAMERA_%03d", ix+1)
		val, set := os.LookupEnv(env)
		if set == true {
			cfg.OccupancyIp = append(cfg.OccupancyIp, val)
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
			cfg.SumIp = append(cfg.SumIp, val)
			cfg.SumUsername = append(cfg.SumUsername, userName)
			cfg.SumPassword = append(cfg.SumPassword, password)
		} else {
			break
		}
	}

	// ensure we have 1 or more endpoints defined
	if len(cfg.OccupancyIp) == 0 && len(cfg.SumIp) == 0 {
		fatalIfError(fmt.Errorf("no camera ip addresses defined"))
	}

	log.Printf("[config] OutQueueName         = [%s]", cfg.OutQueueName)
	log.Printf("[config] MessageBucketName    = [%s]", cfg.MessageBucketName)
	log.Printf("[config] OccupancyQuery       = [%s]", cfg.OccupancyQuery)
	log.Printf("[config] SumQuery             = [%s]", cfg.SumQuery)

	log.Printf("[config] EndpointTimeout      = [%d]", cfg.EndpointTimeout)
	log.Printf("[config] PollTimeSeconds      = [%d]", cfg.PollTimeSeconds)

	for ix, _ := range cfg.OccupancyIp {
		log.Printf("[config] Main camera %03d      = [%s (%s/REDACTED)]", ix+1, cfg.OccupancyIp[ix], cfg.OccupancyUsername[ix])
	}

	for ix, _ := range cfg.SumIp {
		log.Printf("[config] Camera %03d           = [%s (%s/REDACTED)]", ix+1, cfg.SumIp[ix], cfg.SumUsername[ix])
	}

	return &cfg
}

//
// end of file
//
