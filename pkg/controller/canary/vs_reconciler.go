package canary

import (
	"context"
	v1 "github.com/petomalina/krane/pkg/apis/krane/v1"
	"github.com/petomalina/krane/pkg/apis/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/types"
)

// reconcileDestinationRules is an idempotent function that creates/reads the baseline instance
func (r *ReconcileCanary) reconcileVirtualService(ctx context.Context, canary *v1.Canary) (*v1alpha3.DestinationRule, error) {
	L := log.WithValues("canary", canary.Name, "job", "canary")

	// if not testing, we are just gonna skip
	if canary.Status.Progress != v1.CanaryProgress_Canary {
		return nil, nil
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

	return nil, nil
}
