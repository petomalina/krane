package canary

import (
	"context"
	v1 "github.com/petomalina/krane/pkg/apis/krane/v1"
	"github.com/petomalina/krane/pkg/apis/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/types"
)

// reconcileDestinationRules is an idempotent function that creates/reads the baseline instance
func (r *ReconcileCanary) reconcileVirtualService(ctx context.Context, canary *v1.Canary) (*v1alpha3.VirtualService, error) {
	L := log.WithValues("canary", canary.Name, "job", "canary")

	// if not testing, we are just gonna skip
	if canary.Status.Progress != v1.CanaryProgress_Canary {
		return nil, nil
	}

	if canary.Status.Canary.Status == v1.CanaryPhaseStatus_Queued {
		canary.Status.Canary.Status = v1.CanaryPhaseStatus_InProgress
		canary.Status.Canary.Message = "Rerouting In Progress"
		err := r.client.Status().Update(ctx, canary)
		if err != nil {
			return nil, err
		}
	}

	policy, err := r.GetCanaryPolicy(ctx, canary)
	if err != nil {
		return nil, err
	}

	// get the virtual service so we can handle it
	vs := &v1alpha3.VirtualService{}
	err = r.client.Get(ctx, types.NamespacedName{
		Name:      policy.Spec.VirtualService,
		Namespace: canary.Namespace,
	}, vs)
	if err != nil {
		L.Info("An error occured when getting desired VirtualService", "err", err.Error())
		return nil, err
	}

routesLoop:
	for _, httpRoute := range vs.Spec.Http {
		for _, route := range httpRoute.Route {
			// find service with our matching destination
			if route.Destination.Host != policy.Spec.Service {
				continue
			}

			if httpRoute.Route[0].Weight == 0 {
				// either set it to 90 because we'll use 10
				httpRoute.Route[0].Weight = 90
			} else { // or minuts 10 if it's already set
				httpRoute.Route[0].Weight -= 10
			}

			httpRoute.Route = append(httpRoute.Route, &v1alpha3.HTTPRouteDestination{
				Destination: &v1alpha3.Destination{
					Host: canary.Name,
				},
				Weight: 10,
			})

			err = r.client.Update(ctx, vs)
			if err != nil {
				return nil, err
			}

			break routesLoop
		}
	}

	return vs, nil
}
