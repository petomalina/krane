# Krane
Krane is a canary deployment manager for Kubernetes, automatically creating and operating with newly created
canary deployments and reporting back to the status objects.

## Prerequisites

Krane was tested on the following configuration:
- Kubernetes 1.13.5
- Istio 1.1.4 (contained in the `/infra` folder)
  - Grafana and Kiali allowed to see visualized results

## Installation
Installation consists of multiple steps:
- Build the image using `docker build -f ./build/Dockerfile -t <your-name>/<your-repo> .`.
- Push the image to the remote registry by running `docker push <your-name>/<your-repo>`. Note that you may need to use
credentials companions to authenticate when using cloud solutions.
- Add your existing image into the `./deploy/operator.yaml` to override the default operator (by replacing `gcr.io/petomalina/krane`
with `<your-name>/<your-repo>`).
- Installing CRDs via `kubectl apply -f ./deploy/crds` will create Custom Resource Definitions for Canary and CanaryPolicy
objects inside the Kubernetes cluster.
- Installing the operator via direct Kubernetes template configuration using `kubectl apply -f ./deploy`, which will deploy
the Operator workload as well as its Role, RoleBinding and ServiceAccount.

## Building and Running locally
Performing these steps will ensure the projects can be run locally while being connected to a remote cluster:
- Make sure the `kubectl` is connected to your target cluster. This approach differs based on the cloud provider or
when using on-prem solution (e.g. use `gcloud containers` in case you are running on GCP with GKE).
- Install `dep` as a package manager. Note that `vgo` is not supported in `operator-sdk` at the moment and using it
may result in unexpected results (errors by downloading unsupported or unlocked dependencies that `vgo` can't resolve).
- Run `go run ./cmd/manager/main.go`. In case you don't want to use the `default` namespace, use `WATCH_NAMESPACE` environment
variable to override this value.

## Running example
The example uses a single CanaryPolicy object and fully configures the Istio service mesh networking using `Gateway` and `VirtualService`.
You can override any policy to test the features by editing `./example/00-canary-policy.yaml`. Note that the policy will only change for new
canaries and never affects the existing.

Two deployments are part of the example to demonstrate usage on a system:
- `Aggregator` is used as an entrypoint to the system, accepting raw metrics and aggregating them.
- `Storage` is used for temporary storage of the metrics in memory.

Applying Istio infrastructure after the Canary Policy is done by running:
```
kubectl apply -f ./01-gw.yaml
kubectl apply -f ./01-vs.yaml
```

Deploying the application is then done by running:
```
kubectl apply -f ./02-aggregator-deployment.yaml
kubectl apply -f ./02-storage-deployment.yaml
```

After all deployments are reporting as stable, run the canary analysis test by applying:
```
kubectl apply -f ./10-aggregator-canary.yaml
```

This will run the canary deployment for the Aggregator service, and status can be watched using:
```
kubectl get canary -n <namespace> my-app-aggregator-deployment-ah87p9-canary -o yaml -w
```

The `-w` at the end of the command will provide a stream of updates to the canary. Everything can also be observed
from Kiali, Grafana and Prometheus.