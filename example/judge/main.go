package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
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
	margin                              float64

	// start time
	start time.Time
	// difference from the start until this cycle
	diff time.Duration
}

func main() {
	os.Getenv("KRANE_TARGET")
	canaryLabel := os.Getenv("KRANE_CANARY")
	baselineLabel := os.Getenv("KRANE_BASELINE")
	os.Getenv("KRANE_NAMESPACE")
	timeBoundary, _ := time.ParseDuration(os.Getenv("KRANE_BOUNDARY_TIME"))
	fmt.Println("Time Boundary:", timeBoundary, "\nCanary:", canaryLabel, "\nBaseline:", baselineLabel)

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
		fmt.Println("Starting next reconciliation")

		// calculate diff so we can pass it into the reconciliation loop
		diff := time.Now().Sub(start)

		for _, diffMetric := range diffMetrics {
			err := reconcile(ctx, config{
				diffMetric.Metric,
				diffMetric.Container,
				baselineLabel,
				canaryLabel,
				diffMetric.Diff,
				start,
				diff,
			}, prom)
			if err != nil {
				if errors.Cause(err) == ErrMetricFailed {
					fmt.Println(err.Error())
					os.Exit(1)
				}

				fmt.Println("An unexpected error occurred during reconciliation: ", err.Error())
			}
		}

		// break once we are done with the time boundary
		if timeBoundary > 0 && diff >= timeBoundary {
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
	query := newDiffQuery(
		cfg.metric,
		cfg.container,
		cfg.baseline,
		cfg.canary,
	)
	fmt.Println("Executing Query:", query)

	val, err := prom.QueryRange(ctx, query, v1.Range{
		Start: cfg.start,
		End:   time.Now(),
		Step:  time.Second * 5,
	})
	if err != nil {
		return err
	}

	resMatrix := val.(model.Matrix)

	// incidentCounter counts how many times a violation of the metric has occured
	incidentCounter := 0
	for _, dataPoints := range resMatrix {
		if len(dataPoints.Values) <= 0 {
			continue
		}

		for _, val := range dataPoints.Values {
			if float64(val.Value) > cfg.margin {
				incidentCounter++
			} else {
				incidentCounter = 0
			}

			// more than 10 samples of continuous metric violation are erroneous
			if incidentCounter > 10 {
				return errors.Wrap(ErrMetricFailed, "A metric has failed on querying: "+query)
			}
		}
	}

	return nil
}
