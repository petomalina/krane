package canary

import (
	"context"
	"encoding/json"
	v1 "github.com/petomalina/krane/pkg/apis/krane/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// reconcileDestinationRules is an idempotent function that creates/reads the baseline instance
func (r *ReconcileCanary) reconcileJudgeJob(ctx context.Context, canary *v1.Canary) (*corev1.Pod, error) {
	if canary.Status.Progress == v1.CanaryProgress_Cleanup {
		return nil, nil
	}

	L := log.WithValues("canary", canary.Name, "job", "judge")

	// if not testing, we are just gonna skip
	if canary.Status.Progress != v1.CanaryProgress_Testing && canary.Status.Progress != v1.CanaryProgress_Canary {
		return nil, nil
	}

	L.Info("Reconciling step")
	defer L.Info("Reconciliation complete")

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

	// new values for the test job
	var newStatus v1.CanaryPhaseStatus
	var newMessage string

	// update run state if the pod is running
	var runState *corev1.ContainerStateRunning
	if runState = GetContainerRunningState(judgeJob, "judge"); runState != nil {
		newStatus = v1.CanaryPhaseStatus_InProgress
		newMessage = "Job running"
	}

	// update terminal state if the pod has died
	var termState *corev1.ContainerStateTerminated
	if termState = GetContainerTerminalState(judgeJob, "judge"); termState != nil {
		switch termState.ExitCode {
		case 0:
			newStatus = v1.CanaryPhaseStatus_Success
			newMessage = "Job finished successfully"
		default:
			newStatus = v1.CanaryPhaseStatus_Failure
			newMessage = "Job failed"
		}
	}

	if newStatus != canary.Status.Judging.Status || canary.Status.Judging.Message != newMessage {
		canary.Status.Judging.Status = newStatus
		canary.Status.Judging.Message = newMessage

		err = r.client.Status().Update(ctx, canary)
		if err != nil {
			return nil, err
		}
	}

	return judgeJob, nil
}

func GetJudgeJobName(c *v1.Canary) string {
	return c.Name + "-judge"
}

func (r *ReconcileCanary) CreateJudgeJob(canary *v1.Canary, policy *v1.CanaryPolicy) (*corev1.Pod, error) {
	labels := map[string]string{
		"app":     GetJudgeJobName(canary),
		"version": "stable",
	}

	envs := []corev1.EnvVar{}
	if policy.Spec.JudgeSpec.Boundary.Time != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "KRANE_BOUNDARY_TIME",
			Value: policy.Spec.JudgeSpec.Boundary.Time,
		})
	}

	if len(policy.Spec.JudgeSpec.DiffMetrics) > 0 {
		mm, _ := json.Marshal(policy.Spec.JudgeSpec.DiffMetrics)
		envs = append(envs, corev1.EnvVar{
			Name:  "KRANE_DIFF_METRICS",
			Value: string(mm),
		})
	}

	if len(policy.Spec.JudgeSpec.ThresholdMetrics) > 0 {
		mm, _ := json.Marshal(policy.Spec.JudgeSpec.DiffMetrics)

		envs = append(envs, corev1.EnvVar{
			Name:  "KRANE_THRESHOLD_METRICS",
			Value: string(mm),
		})
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetJudgeJobName(canary),
			Namespace: canary.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			// TODO: allow modification by user
			Containers: []corev1.Container{
				{
					Name:            "judge",
					Image:           policy.Spec.JudgeSpec.Image,
					ImagePullPolicy: corev1.PullAlways,
					Command:         policy.Spec.JudgeSpec.Cmd,
					Env: append([]corev1.EnvVar{
						{
							// canary services have the same name as canaries themselves
							Name:  "KRANE_TARGET",
							Value: canary.Name,
						},
						{
							Name:  "KRANE_PROMETHEUS",
							Value: "http://prometheus.istio-system.svc.cluster.local:9090",
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
					}, envs...),
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
