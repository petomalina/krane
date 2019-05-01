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
	for {
		_, err := http.Post("http://"+target+"/observe?metric=hits&value=1", "", nil)
		if err != nil {
			fmt.Println("WARN: an error occured when hitting the service", err)
			os.Exit(1)
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
	}

	fmt.Println("Total Requests: ", reqs)
}
