package deployment

import (
	"context"
	v1 "github.com/petomalina/krane/pkg/apis/krane/v1"
	appsv1 "k8s.io/api/apps/v1"
)

func (r *ReconcileDeployment) reconcileCanaryDeployment(ctx context.Context, canaryInstance *appsv1.Deployment, policy *v1.CanaryPolicy) error {
	// skip the update if we already have labels
	if canaryInstance.Spec.Template.ObjectMeta.Labels[CanaryConfigLabel] != "" && canaryInstance.Spec.Template.ObjectMeta.Labels[CanaryPolicyLabel] != "" {
		return nil
	}

	canaryInstance.Spec.Template.ObjectMeta.Labels[CanaryConfigLabel] = MakeCanaryConfigName(policy, canaryInstance)
	canaryInstance.Spec.Template.ObjectMeta.Labels[CanaryPolicyLabel] = canaryInstance.Labels[CanaryPolicyLabel]

	err := r.client.Update(ctx, canaryInstance)
	if err != nil {
		return err
	}

	return nil
}
