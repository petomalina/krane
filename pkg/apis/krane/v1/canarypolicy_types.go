package v1

import (
	"github.com/petomalina/krane/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BaselineMode string

const (
	BaselineModeNew BaselineMode = "NEW"
	BaselineModeOld              = "OLD"
)

// CanaryPolicySpec defines the desired state of CanaryPolicy
// +k8s:openapi-gen=true
type CanaryPolicySpec struct {
	// Base is the original deployment that the canary will use
	// to copy baseline
	Base string `json:"base,omitempty"`
	// VirtualService is a name of bindable virtualservice that
	// will be used for traffic splitting
	VirtualService string `json:"virtualService,omitempty"`
	// Mode is used to determine if baseline uses original
	// or canary configuration
	BaselineMode BaselineMode `json:"baselineMode,omitempty"`
	// DestinationRule is a rule that will be created when a new
	// canary will be deployed
	DestinationRule v1alpha3.DestinationRuleSpec
}

// CanaryPolicyStatus defines the observed state of CanaryPolicy
// +k8s:openapi-gen=true
type CanaryPolicyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CanaryPolicy is the Schema for the canarypolicies API
// +k8s:openapi-gen=true
type CanaryPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CanaryPolicySpec   `json:"spec,omitempty"`
	Status CanaryPolicyStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CanaryPolicyList contains a list of CanaryPolicy
type CanaryPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CanaryPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CanaryPolicy{}, &CanaryPolicyList{})
}
