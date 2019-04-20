package deployment

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
)

const (
	CanaryPolicyLabel = "krane.sh/canary-policy"
	CanaryConfigLabel = "krane.sh/canary-config"
)

// CanaryObjectPredicate filters out all Deployments that are not Canaries
type CanaryObjectPredicate struct{}

func (p *CanaryObjectPredicate) Create(e event.CreateEvent) bool {
	if !hasCanaryPolicyLabel(e.Meta.GetLabels()) {
		return false
	}

	return true
}

func (p *CanaryObjectPredicate) Delete(e event.DeleteEvent) bool {
	return false
}

func (p *CanaryObjectPredicate) Update(e event.UpdateEvent) bool {
	if !hasCanaryPolicyLabel(e.MetaNew.GetLabels()) {
		return false
	}

	return true
}

func (p *CanaryObjectPredicate) Generic(e event.GenericEvent) bool {
	if !hasCanaryPolicyLabel(e.Meta.GetLabels()) {
		return false
	}

	return true

}

func hasCanaryPolicyLabel(labels map[string]string) bool {
	_, ok := labels[CanaryPolicyLabel]
	return ok
}
