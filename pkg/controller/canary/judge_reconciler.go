package canary

import (
	"context"
	v1 "github.com/petomalina/krane/pkg/apis/krane/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileDestinationRules is an idempotent function that creates/reads the baseline instance
func (r *ReconcileCanary) reconcileJudgeJob(ctx context.Context, canary *v1.Canary) (*corev1.Pod, error) {
	L := log.WithValues("canary", canary.Name, "job", "judge")

	// if not testing, we are just gonna skip
	if canary.Status.Progress != v1.CanaryProgress_Testing && canary.Status.Progress != v1.CanaryProgress_Canary {
		return nil, nil
	}

	L.Info("Reconciling step")

	policy, err := r.GetCanaryPolicy(ctx, canary)
	if err != nil {
		return nil, err
	}

	// get the test pod so we can handle it
	judgeJob := &corev1.Pod{}
	err = r.client.Get(ctx, types.NamespacedName{
		Name:      GetJudgeJobName(canary),
		Namespace: canary.Namespace,
	}, judgeJob)
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}

		judgeJob, err = r.CreateJudgeJob(canary, policy)
		if err != nil {
			return nil, err
		}

		err = r.client.Create(ctx, judgeJob)
		if err != nil {
			return nil, err
		}

		canary.Status.Judging.Status = v1.CanaryPhaseStatus_InProgress
		canary.Status.Judging.Message = "Job initialized"
		canary.Status.Judging.PodName = judgeJob.Name
		err = r.client.Status().Update(ctx, canary)
		if err != nil {
			return nil, err
		}
	}

	L.Info("Reconciliation complete")

	return judgeJob, nil
}

func GetJudgeJobName(c *v1.Canary) string {
	return c.Name + "-judge"
}

func (r *ReconcileCanary) CreateJudgeJob(canary *v1.Canary, policy *v1.CanaryPolicy) (*corev1.Pod, error) {
	labels := map[string]string{}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetJudgeJobName(canary),
			Namespace: canary.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "judge",
					Image:   policy.Spec.JudgeSpec.Image,
					Command: policy.Spec.JudgeSpec.Cmd,
					Env: []corev1.EnvVar{
						{
							// canary services have the same name as canaries themselves
							Name:  "KRANE_TARGET",
							Value: canary.Name,
						},
						{
							Name:  "KRANE_CANARY",
							Value: canary.Spec.Deployments.Canary,
						},
						{
							Name:  "KRANE_BASELINE",
							Value: canary.Spec.Deployments.Baseline,
						},
						{
							Name:  "KRANE_NAMESPACE",
							Value: canary.Namespace,
						},
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}

	if err := controllerutil.SetControllerReference(canary, pod, r.scheme); err != nil {
		return nil, err
	}

	return pod, nil
}
