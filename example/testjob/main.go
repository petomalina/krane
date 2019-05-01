package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	reqLimit, _ := strconv.Atoi(os.Getenv("KRANE_BOUNDARY_REQUESTS"))
	timeLimit, _ := time.ParseDuration(os.Getenv("KRANE_BOUNDARY_TIME"))
	target := os.Getenv("KRANE_TARGET")

	reqs := 0
	start := time.Now()

	retries := 100
	for {
		_, err := http.Post("http://"+target+"/observe?metric=hits&value=1", "", nil)
		if err != nil {
			// we want to wait for the istio proxy to start before sending any traffic
			if retries == 0 {
				os.Exit(0)
			}

			retries--
			fmt.Println("WARN: an error occured when hitting the service", err)
			time.Sleep(time.Millisecond * 500)
		}

		diff := time.Now().Sub(start)
		reqs++

		// break on request boundary
		if reqLimit != 0 && reqLimit >= reqs {
			break
		}

		// break on time boundary
		if timeLimit != 0 && timeLimit >= diff {
			break
		}

		time.Sleep(time.Millisecond * 200)
	}

	fmt.Println("Total Requests: ", reqs)
}
