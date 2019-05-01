package main

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"os"
	"time"
)

var (
	q = "container_cpu_system_seconds_total{container_name=\"aggregator\",name=~\".*canary.*\"} - on (container_name) container_cpu_system_seconds_total{container_name=\"aggregator\",name=~\".*baseline.*\"}"
)

func main() {
	os.Getenv("KRANE_TARGET")
	os.Getenv("KRANE_CANARY")
	os.Getenv("KRANE_BASELINE")
	os.Getenv("KRANE_NAMESPACE")

	os.Getenv("KRANE_DIFF_METRICS")
	os.Getenv("KRANE_THRESHOLD_METRICS")

	promCfg := api.Config{Address: os.Getenv("KRANE_PROMETHEUS")}
	promClient, err := api.NewClient(promCfg)
	if err != nil {
		panic(err)
	}

	prom := v1.NewAPI(promClient)

	ctx := context.Background()
	value, err := prom.QueryRange(ctx, newDiffQuery(
		"container_cpu_system_seconds_total",
		"aggregator",
		"aggregator-deployment-baseline",
		"aggregator-deployment-canary",
	), v1.Range{
		Start: time.Now().Add(-time.Minute * 20),
		End:   time.Now(),
		Step:  time.Second * 5,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(value)
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
