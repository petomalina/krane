package deployment

import (
	"context"
	v1 "github.com/petomalina/krane/pkg/apis/krane/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ReconcileDeployment) reconcileCanaryConfig(ctx context.Context, canaryInstance *appsv1.Deployment, policy *v1.CanaryPolicy) (*v1.Canary, error) {
	canaryConfig := &v1.Canary{}

	err := r.client.Get(ctx, types.NamespacedName{
		Name:      MakeCanaryConfigName(policy, canaryInstance),
		Namespace: policy.Name,
	}, canaryConfig)
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}

		err := r.client.Create(ctx, r.createCanaryConfig(ctx, canaryInstance, policy))
		if err != nil {
			return nil, err
		}
	}

	return canaryConfig, nil
}

func (r *ReconcileDeployment) createCanaryConfig(ctx context.Context, canaryInstance *appsv1.Deployment, policy *v1.CanaryPolicy) *v1.Canary {
	return &v1.Canary{
		ObjectMeta: metav1.ObjectMeta{
			Name:      policy.Name + "-" + canaryInstance.Name,
			Namespace: policy.Namespace,
		},
		Spec: v1.CanarySpec{
			Policy:   policy.Name,
			Canary:   canaryInstance.Name,
			Baseline: canaryInstance.Name + "-baseline",
			Base:     policy.Spec.Base,
		},
		Status: v1.CanaryStatus{
			Progress: v1.CanaryProgress_Initializing,
		},
	}
}
