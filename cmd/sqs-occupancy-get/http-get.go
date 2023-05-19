package main

import (
	"fmt"
	"github.com/icholy/digest"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

var maxHttpRetries = 3
var retrySleepTime = 100 * time.Millisecond

func newDigestClient(username string, password string, timeout int) *http.Client {

	return &http.Client{
		Transport: &digest.Transport{
			//MaxIdleConnsPerHost: 1,
			Username: username,
			Password: password,
		},
		Timeout: time.Duration(timeout) * time.Second,
	}
}

func httpGet(workerId int, url string, client *http.Client) ([]byte, error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("ERROR: worker %d GET %s failed with error (%s)", workerId, url, err)
		return nil, err
	}

	var response *http.Response
	count := 0
	for {
		start := time.Now()
		response, err = client.Do(req)
		duration := time.Since(start)
		log.Printf("INFO: worker %d GET %s (elapsed %d ms)", workerId, url, duration.Milliseconds())

		count++
		if err != nil {
			if canRetry(err) == false {
				log.Printf("ERROR: worker %d GET %s failed with error (%s)", workerId, url, err)
				return nil, err
			}

			// break when tried too many times
			if count >= maxHttpRetries {
				log.Printf("ERROR: worker %d GET %s failed with error (%s), retry # %d, giving up", workerId, url, err, count)
				return nil, err
			}

			log.Printf("WARNING: worker %d GET %s failed with error (%s), retry # %d", workerId, url, err, count)

			// sleep for a bit before retrying
			time.Sleep(retrySleepTime)
		} else {

			defer response.Body.Close()

			if response.StatusCode != http.StatusOK {

				log.Printf("ERROR: worker %d GET %s failed with status %d", workerId, url, response.StatusCode)

				body, _ := ioutil.ReadAll(response.Body)

				return body, fmt.Errorf("request returns HTTP %d", response.StatusCode)
			} else {
				body, err := ioutil.ReadAll(response.Body)
				if err != nil {
					log.Printf("ERROR: worker %d reading response body (%s)", workerId, err.Error())
					return nil, err
				}

				//log.Printf( body )
				return body, nil
			}
		}
	}
}

// examines the error and decides if it can be retried
func canRetry(err error) bool {

	if strings.Contains(err.Error(), "operation timed out") == true {
		return true
	}

	if strings.Contains(err.Error(), "Client.Timeout exceeded") == true {
		return true
	}

	if strings.Contains(err.Error(), "write: broken pipe") == true {
		return true
	}

	if strings.Contains(err.Error(), "no such host") == true {
		return true
	}

	if strings.Contains(err.Error(), "network is down") == true {
		return true
	}

	return false
}

//
// end of file
//
