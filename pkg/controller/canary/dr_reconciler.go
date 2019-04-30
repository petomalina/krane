package canary

import (
	"context"
	v1 "github.com/petomalina/krane/pkg/apis/krane/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/petomalina/krane/pkg/apis/networking/v1alpha3"
	"github.com/petomalina/krane/pkg/controller/deployment"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// reconcileDestinationRules is an idempotent function that creates/reads the baseline instance
func (r *ReconcileCanary) reconcileDestinationRules(ctx context.Context, canary *v1.Canary) (*v1alpha3.DestinationRule, error) {
	// get the policy
	policy := &v1.CanaryPolicy{}
	err := r.client.Get(ctx, types.NamespacedName{
		Name:      canary.Spec.Policy,
		Namespace: canary.Namespace,
	}, policy)
	if err != nil {
		return nil, err
	}

	// check for baseline deployment
	dr := &v1alpha3.DestinationRule{}
	err = r.client.Get(ctx, types.NamespacedName{
		Namespace: canary.Namespace,
		Name:      MakeDestinationRuleName(canary),
	}, dr)
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}

		dr, err = r.createBaselineDestinationRule(ctx, canary, policy)
		if err != nil {
			return nil, err
		}

		err = r.client.Create(ctx, dr)
		if err != nil {
			return nil, err
		}
	}

	return dr, nil
}

func (r *ReconcileCanary) createBaselineDestinationRule(ctx context.Context, canary *v1.Canary, policy *v1.CanaryPolicy) (*v1alpha3.DestinationRule, error) {
	defaultDr := policy.Spec.DestinationRule
	// set Subset and Host as we don't want people to override this
	defaultDr.Subsets = []*v1alpha3.Subset{
		{
			Name: canary.Name,
			Labels: map[string]string{
				deployment.CanaryConfigLabel: canary.Name,
				deployment.CanaryPolicyLabel: canary.Spec.Policy,
			},
		},
	}
	defaultDr.Host = policy.Name + "-" + canary.Spec.Deployments.Canary

	dr := &v1alpha3.DestinationRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      MakeDestinationRuleName(canary),
			Namespace: canary.Namespace,
			Labels: map[string]string{
				deployment.CanaryConfigLabel: canary.Name,
				deployment.CanaryPolicyLabel: canary.Spec.Policy,
			},
		},
		Spec: defaultDr,
	}

	err := controllerutil.SetControllerReference(canary, dr, r.scheme)
	if err != nil {
		return nil, err
	}

	return dr, nil
}

type DestinationRuleType string

func MakeDestinationRuleName(canary *v1.Canary) string {
	return canary.Name
}
