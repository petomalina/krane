package canary

import (
	"context"
	v1 "github.com/petomalina/krane/pkg/apis/krane/v1"
)

func (r *ReconcileCanary) reconcileReporting(ctx context.Context, canary *v1.Canary) error {
	if canary.Status.Progress != v1.CanaryProgress_Reporting {
		return nil
	}

	// already done, don't requeue
	if canary.Status.Reporting.Status == v1.CanaryPhaseStatus_Success {
		return nil
	}

	canary.Status.Reporting.Status = v1.CanaryPhaseStatus_Success
	canary.Status.Reporting.Message = "Reported"
	err := r.client.Status().Update(ctx, canary)
	if err != nil {
		return nil
	}

	return nil
}
