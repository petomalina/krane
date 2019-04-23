package deployment

import (
	"context"
	"github.com/petomalina/krane/pkg/apis/krane/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ReconcileDeployment) reconcileCanaryService(ctx context.Context, canaryInstance *appsv1.Deployment, policy *v1.CanaryPolicy) (*corev1.Service, error) {
	svc := &corev1.Service{}

	err := r.client.Get(ctx, types.NamespacedName{
		Name:      MakeCanaryServiceName(policy, canaryInstance),
		Namespace: policy.Namespace,
	}, svc)
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}

		svc, err = r.createCanaryService(ctx, canaryInstance, policy)
		if err != nil {
			return nil, err
		}

		err = r.client.Create(ctx, svc)
		if err != nil {
			return nil, err
		}
	}

	return svc, nil
}

func (r *ReconcileDeployment) createCanaryService(ctx context.Context, canaryInstance *appsv1.Deployment, policy *v1.CanaryPolicy) (*corev1.Service, error) {
	if len(policy.Spec.Ports) <= 0 {
		policy.Spec.Ports = []corev1.ServicePort{
			{
				Port:     80,
				Name:     "default",
				Protocol: corev1.ProtocolTCP,
			},
		}
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      MakeCanaryServiceName(policy, canaryInstance),
			Namespace: canaryInstance.Namespace,
			Labels: map[string]string{
				CanaryConfigLabel: MakeCanaryConfigName(policy, canaryInstance),
				CanaryPolicyLabel: canaryInstance.Labels[CanaryPolicyLabel],
			},
		},
		Spec: corev1.ServiceSpec{
			// set selector to all deployments labeled with this config name and policy
			Selector: map[string]string{
				CanaryConfigLabel: MakeCanaryConfigName(policy, canaryInstance),
				CanaryPolicyLabel: canaryInstance.Labels[CanaryPolicyLabel],
			},
			// copy over ports defined by users
			Ports: policy.Spec.Ports,
		},
	}

	err := controllerutil.SetControllerReference(canaryInstance, svc, r.scheme)
	if err != nil {
		return nil, err
	}

	return svc, nil
}

func MakeCanaryServiceName(p *v1.CanaryPolicy, c *appsv1.Deployment) string {
	return p.Name + "-" + c.Name
}
