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

	Endpoints       []string // the list of endpoints
	Username        []string // the list of endpoint usernames
	Password        []string // the list of endpoint passwords
	EndpointTimeout int      // the endpoint timeout in seconds
	PollTimeSeconds int      // the endpoint poll time in seconds
}

func envWithDefault(env string, defaultValue string) string {
	val, set := os.LookupEnv(env)

	if set == false {
		log.Printf("DEBUG: environment variable not set: [%s] using default value [%s]", env, defaultValue)
		return defaultValue
	}

	return val
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
	cfg.EndpointTimeout = envToInt("SQS_OCCUPANCY_GET_ENDPOINT_TIMEOUT")
	cfg.PollTimeSeconds = envToInt("SQS_OCCUPANCY_GET_POLL_IN_SECONDS")
	userName := ensureSetAndNonEmpty("SQS_OCCUPANCY_GET_USERNAME")
	password := ensureSetAndNonEmpty("SQS_OCCUPANCY_GET_PASSWORD")

	for ix := 0; ix < maxCameras; ix++ {
		env := fmt.Sprintf("SQS_OCCUPANCY_GET_ENDPOINT_%03d", ix+1)
		val, set := os.LookupEnv(env)
		if set == true {
			cfg.Endpoints = append(cfg.Endpoints, val)
			cfg.Username = append(cfg.Username, userName)
			cfg.Password = append(cfg.Password, password)
		} else {
			break
		}
	}

	// ensure we have 1 or more endpoints defined
	if len(cfg.Endpoints) == 0 {
		fatalIfError(fmt.Errorf("no camera endpoints defined. Specify using the environment variable SQS_OCCUPANCY_GET_ENDPOINT_nnn"))
	}

	log.Printf("[CONFIG] OutQueueName         = [%s]", cfg.OutQueueName)
	log.Printf("[CONFIG] MessageBucketName    = [%s]", cfg.MessageBucketName)
	log.Printf("[CONFIG] EndpointTimeout      = [%d]", cfg.EndpointTimeout)
	log.Printf("[CONFIG] PollTimeSeconds      = [%d]", cfg.PollTimeSeconds)

	for ix, _ := range cfg.Endpoints {
		log.Printf("[CONFIG] Endpoint %03d         = [%s (%s/REDACTED)]", ix+1, cfg.Endpoints[ix], cfg.Username[ix])
	}

	return &cfg
}
