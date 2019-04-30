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
func (r *ReconcileCanary) reconcileTestJob(ctx context.Context, canary *v1.Canary) (*corev1.Pod, error) {
	L := log.WithValues("canary", canary.Name, "job", "test")

	// if not testing, we are just gonna skip
	if canary.Status.Progress != v1.CanaryProgress_Testing {
		return nil, nil
	}

	L.Info("Reconciling test step")
	defer L.Info("Reconciliation complete")

	policy, err := r.GetCanaryPolicy(ctx, canary)
	if err != nil {
		return nil, err
	}

	// get the test pod so we can handle it
	testJob := &corev1.Pod{}
	err = r.client.Get(ctx, types.NamespacedName{
		Name:      GetTestJobname(canary),
		Namespace: canary.Namespace,
	}, testJob)
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}

		testJob, err = r.CreateTestJob(canary, policy)
		if err != nil {
			return nil, err
		}

		err = r.client.Create(ctx, testJob)
		if err != nil {
			return nil, err
		}

		canary.Status.Testing.Status = v1.CanaryPhaseStatus_InProgress
		canary.Status.Testing.Message = "Job initialized"
		canary.Status.Testing.PodName = testJob.Name
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
	if runState = GetContainerRunningState(testJob, "testjob"); runState != nil {
		newStatus = v1.CanaryPhaseStatus_InProgress
		newMessage = "Job running"
	}

	// update terminal state if the pod has died
	var termState *corev1.ContainerStateTerminated
	if termState = GetContainerTerminalState(testJob, "testjob"); termState != nil {
		switch termState.ExitCode {
		case 0:
			newStatus = v1.CanaryPhaseStatus_Success
			newMessage = "Job finished successfully"
		default:
			newStatus = v1.CanaryPhaseStatus_Failure
			newMessage = "Job failed"
		}
	}

	if newStatus != canary.Status.Testing.Status || canary.Status.Testing.Message != newMessage {
		canary.Status.Testing.Status = newStatus
		canary.Status.Testing.Message = newMessage

		err = r.client.Status().Update(ctx, canary)
		if err != nil {
			return nil, err
		}
	}

	return testJob, nil
}

func GetTestJobname(c *v1.Canary) string {
	return c.Name + "-testjob"
}

func (r *ReconcileCanary) CreateTestJob(canary *v1.Canary, policy *v1.CanaryPolicy) (*corev1.Pod, error) {
	labels := map[string]string{}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetTestJobname(canary),
			Namespace: canary.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "testjob",
					Image:   policy.Spec.TestSpec.Image,
					Command: policy.Spec.TestSpec.Cmd,
					Env: []corev1.EnvVar{
						{
							// canary services have the same name as canaries themselves
							Name:  "KRANE_TARGET",
							Value: canary.Name,
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

// IsPodDone returns true if the pod has all containers completed
func GetContainerTerminalState(job *corev1.Pod, containerName string) *corev1.ContainerStateTerminated {
	for _, c := range job.Status.ContainerStatuses {
		// skip the istio-proxy container so we can ignore it
		if c.Name == containerName && c.State.Terminated != nil {
			return c.State.Terminated
		}

	}

	return nil
}

// GetContainerRunningState returns a running state if the container has that
func GetContainerRunningState(job *corev1.Pod, containerName string) *corev1.ContainerStateRunning {
	for _, c := range job.Status.ContainerStatuses {
		// skip the istio-proxy container so we can ignore it
		if c.Name == containerName && c.State.Terminated != nil {
			return c.State.Running
		}
	}

	return nil
}
