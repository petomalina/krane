package deployment

import (
	"context"
	v1 "github.com/petomalina/krane/pkg/apis/krane/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// reconcileBaseline is an idempotent function that creates/reads the baseline instance
func (r *ReconcileDeployment) reconcileBaseline(ctx context.Context, canaryInstance *appsv1.Deployment, policy *v1.CanaryPolicy) (*appsv1.Deployment, error) {
	// check for baseline deployment
	baseline := &appsv1.Deployment{}
	err := r.client.Get(ctx, types.NamespacedName{
		Namespace: canaryInstance.Namespace,
		Name:      MakeBaselineName(canaryInstance),
	}, baseline)
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}

		baseline, err = r.createBaselineDeployment(ctx, canaryInstance, policy)
		if err != nil {
			return nil, err
		}

		err = r.client.Create(ctx, baseline)
		if err != nil {
			return nil, err
		}
	}

	return baseline, nil
}

func (r *ReconcileDeployment) createBaselineDeployment(ctx context.Context, canaryInstance *appsv1.Deployment, policy *v1.CanaryPolicy) (*appsv1.Deployment, error) {
	baseline := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      canaryInstance.Name + "-baseline",
			Namespace: canaryInstance.Namespace,
			Labels: map[string]string{
				"release": "baseline",
			},
		},
	}

	// get the old deployment to retrieve containers
	base := &appsv1.Deployment{}
	err := r.client.Get(ctx, types.NamespacedName{
		Namespace: canaryInstance.Namespace,
		Name:      policy.Spec.Base,
	}, base)
	if err != nil {
		return nil, err
	}

	// Default to new baseline mode
	if policy.Spec.BaselineMode == "" {
		policy.Spec.BaselineMode = v1.BaselineModeNew
	}

	switch policy.Spec.BaselineMode {
	case v1.BaselineModeNew:
		baseline.Spec = *canaryInstance.Spec.DeepCopy()
		// copy over previous container configuration
		baseline.Spec.Template.Spec.Containers = base.Spec.Template.Spec.Containers

	case v1.BaselineModeOld:
		baseline.Spec = *base.Spec.DeepCopy()
	}

	// argh, golang, why you no support pointers
	singleReplica := int32(1)
	baseline.Spec.Replicas = &singleReplica

	// connect selectors
	baseline.Spec.Selector.MatchLabels["release"] = "baseline"
	baseline.Spec.Template.ObjectMeta.Labels["release"] = "baseline"

	return baseline, nil
}
