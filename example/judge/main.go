package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"os"
	"time"
)

var (
	q = "container_cpu_system_seconds_total{container_name=\"aggregator\",name=~\".*canary.*\"} - on (container_name) container_cpu_system_seconds_total{container_name=\"aggregator\",name=~\".*baseline.*\"}"
)

type ThresholdMetric struct {
	Metric string `json:"metric,omitempty"`
	Value  string `json:"value,omitempty"`
}

type DiffMetric struct {
	Metric    string  `json:"metric,omitempty"`
	Container string  `json:"container,omitempty"`
	Diff      float64 `json:"diff,omitempty"`
}

var (
	ErrMetricFailed = errors.New("A metric failed to succeed in the test")
)

type config struct {
	metric, container, baseline, canary string
	// start time
	start time.Time
	// difference from the start until this cycle
	diff time.Duration
}

func main() {
	os.Getenv("KRANE_TARGET")
	os.Getenv("KRANE_CANARY")
	os.Getenv("KRANE_BASELINE")
	os.Getenv("KRANE_NAMESPACE")
	timeBoundary, _ := time.ParseDuration(os.Getenv("KRANE_BOUNDARY_TIME"))
	fmt.Println("Time Boundary:", timeBoundary)

	// parse diff and threshold metrics
	diffMetrics := []DiffMetric{}
	thresholdMetrics := []ThresholdMetric{}

	if diffs := os.Getenv("KRANE_DIFF_METRICS"); diffs != "" {
		err := json.Unmarshal([]byte(diffs), &diffMetrics)
		if err != nil {
			panic(err)
		}
	}

	if thresholds := os.Getenv("KRANE_THRESHOLD_METRICS"); thresholds != "" {
		err := json.Unmarshal([]byte(thresholds), &thresholdMetrics)
		if err != nil {
			panic(err)
		}
	}

	// wait for the istio proxy
	time.Sleep(time.Second * 10)

	promCfg := api.Config{Address: os.Getenv("KRANE_PROMETHEUS")}
	promClient, err := api.NewClient(promCfg)
	if err != nil {
		panic(err)
	}

	prom := v1.NewAPI(promClient)
	ctx := context.Background()

	start := time.Now()
	for {
		// calculate diff so we can pass it into the reconciliation loop
		diff := time.Now().Sub(start)

		err := reconcile(ctx, config{
			"container_cpu_system_seconds_total",
			"aggregator",
			"aggregator-deployment-baseline",
			"aggregator-deployment-canary",
			start,
			diff,
		}, prom)
		if err != nil {
			if err == ErrMetricFailed {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			fmt.Println("An unexpected error occurred during reconciliation: ", err.Error())
		}

		// break once we are done with the time boundary
		if diff >= timeBoundary {
			fmt.Println("Time boundary impact, breaking reconciliation")
			break
		}

		time.Sleep(time.Second * 20)
	}

	fmt.Println("Judging done")
}

// e.g.
// metric=container_cpu_system_seconds_total
// containerName=aggregator
// baseline= ...
// canary= ...
func newDiffQuery(metric, containerName, baseline, canary string) string {
	return metric + "{container_name=\"" + containerName + "\",name=~\".*" + canary +
		".*\"} - on (container_name) " +
		metric + "{container_name=\"" + containerName + "\",name=~\".*" + baseline + ".*\"}"
}

func reconcile(ctx context.Context, cfg config, prom v1.API) error {
	_, err := prom.QueryRange(ctx, newDiffQuery(
		cfg.metric,
		cfg.container,
		cfg.baseline,
		cfg.canary,
	), v1.Range{
		Start: cfg.start,
		End:   time.Now(),
		Step:  time.Second * 5,
	})
	if err != nil {
		return err
	}

	return nil
}
