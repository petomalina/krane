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
		canary.Status.Canary.Message = "Rerouting Configuration In Progress"
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

	httpIndex, routeIndex := r.FindMatchingHttpAndRoute(policy.Spec.Service, vs)
	if httpIndex == -1 || routeIndex == -1 {
		L.Info("Could not find the matching service, waiting for the VS configuration")
		return nil, nil
	}

	canaryRoute := r.FindMatchingRouteInHttpRoute(canary.Name, vs.Spec.Http[httpIndex])
	if canaryRoute != -1 {
		// the canary status update will be delayed but always updated in case we are already configured and in progress
		if canary.Status.Canary.Status == v1.CanaryPhaseStatus_InProgress {
			canary.Status.Canary.Status = v1.CanaryPhaseStatus_Success
			canary.Status.Canary.Message = "Rerouting Configuration In Progress"
			err = r.client.Status().Update(ctx, canary)
			if err != nil {
				return nil, err
			}
		}

		L.Info("Canary route already configured, skipping VirtualService configuration", "VirtualService", vs.Name)
		return nil, nil
	}

	L.Info("VS", "VirtualService", vs)

	baseDestination := vs.Spec.Http[httpIndex].Route[routeIndex]
	if baseDestination.Weight == 0 {
		// either set it to 90 because we'll use 10
		baseDestination.Weight = 90
	} else { // or minus 10 if it's already set
		baseDestination.Weight -= 10
	}

	vs.Spec.Http[httpIndex].Route = append(vs.Spec.Http[httpIndex].Route, &v1alpha3.HTTPRouteDestination{
		Destination: &v1alpha3.Destination{
			Host: canary.Name,
		},
		Weight: 10,
	})

	L.Info("Updating the target virtualservice", "VirtualService", vs.Name)
	err = r.client.Update(ctx, vs)
	if err != nil {
		return nil, err
	}

	return vs, nil
}

func (r *ReconcileCanary) FindMatchingHttpAndRoute(svc string, vs *v1alpha3.VirtualService) (int, int) {
	for httpIndex, httpRoute := range vs.Spec.Http {
		for routeIndex, route := range httpRoute.Route {
			if route.Destination.Host == svc {
				return httpIndex, routeIndex
			}
		}
	}

	return -1, -1
}

func (r *ReconcileCanary) FindMatchingRouteInHttpRoute(svc string, http *v1alpha3.HTTPRoute) int {
	for routeIndex, route := range http.Route {
		if route.Destination.Host == svc {
			return routeIndex
		}
	}

	return -1
}
