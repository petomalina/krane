package canary

import (
	"context"
	"github.com/petomalina/krane/pkg/apis/krane/v1"
	"github.com/petomalina/krane/pkg/apis/networking/v1alpha3"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (r *ReconcileCanary) reconcileCleanup(ctx context.Context, canary *v1.Canary) error {
	if canary.Status.Progress != v1.CanaryProgress_Cleanup {
		return nil
	}

	// already cleaned up, don't requeue
	if canary.Status.Cleanup.Status == v1.CanaryPhaseStatus_Success {
		return nil
	}

	policy, err := r.GetCanaryPolicy(ctx, canary)
	if err != nil {
		return err
	}

	err = r.cleanupPod(ctx, GetTestJobname(canary), canary.Namespace)
	if err != nil {
		return err
	}

	err = r.cleanupPod(ctx, GetJudgeJobName(canary), canary.Namespace)
	if err != nil {
		return err
	}

	err = r.cleanupDeployment(ctx, canary.Spec.Deployments.Canary, canary.Namespace)
	if err != nil {
		return err
	}

	// cleanup the virtual service weights
	vs := &v1alpha3.VirtualService{}
	err = r.client.Get(ctx, types.NamespacedName{
		Name:      policy.Spec.VirtualService,
		Namespace: canary.Namespace,
	}, vs)
	if err != nil {
		return err
	}

	// find the route matching index ... if we can't find it, we already did this cleanup before
	httpIndex, routeIndex := r.FindMatchingHttpAndRoute(policy.Spec.Service, vs)
	if httpIndex != -1 && routeIndex != -1 {
		canaryRoute := r.FindMatchingRouteInHttpRoute(canary.Name, vs.Spec.Http[httpIndex])
		if canaryRoute != -1 {
			baseDestination := vs.Spec.Http[httpIndex].Route[routeIndex]

			baseDestination.Weight += vs.Spec.Http[httpIndex].Route[canaryRoute].Weight

			// delete the route from remaining routes
			vs.Spec.Http[httpIndex].Route = append(vs.Spec.Http[httpIndex].Route[:canaryRoute], vs.Spec.Http[httpIndex].Route[canaryRoute+1:]...)

			err = r.client.Update(ctx, vs)
			if err != nil {
				return err
			}
		}
	}

	canary.Status.Cleanup.Status = v1.CanaryPhaseStatus_Success
	canary.Status.Cleanup.Message = "Rerouting Configuration In Progress"
	err = r.client.Status().Update(ctx, canary)
	if err != nil {
		return nil
	}

	return nil
}

func (r *ReconcileCanary) cleanupPod(ctx context.Context, name, namespace string) error {
	pod := &corev1.Pod{}
	err := r.client.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, pod)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return r.client.Delete(ctx, pod)
}

func (r *ReconcileCanary) cleanupDeployment(ctx context.Context, name, namespace string) error {
	app := &appsv1.Deployment{}
	err := r.client.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, app)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	return r.client.Delete(ctx, app)
}
