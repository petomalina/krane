package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CanarySpec defines the desired state of Canary
// +k8s:openapi-gen=true
type CanarySpec struct {
	Policy string `json:"policy,omitempty"`

	Canary   string `json:"canary,omitempty"`
	Baseline string `json:"baseline,omitempty"`
	Base     string `json:"base,omitempty"`
}

// CanaryProgress is a progress of the Krane object during its deployment
type CanaryProgress string

const (
	// Initializing indicates that deployment creation and pod initialization
	// is in progress and the operator is waiting until pretest can start
	CanaryProgress_Initializing CanaryProgress = "initializing"
	// Test indicates that the service is being tested using the pretest
	// container to make sure the container can be routed
	CanaryProgress_Test = "pretest"
	// Canary indicates that the operator has split the traffic, created an
	// additional replica of the stable Krane deployment and started the collection
	// of prometheus metrics
	CanaryProgress_Canary = "canary"
	// Judging indicates that the Judge container has started to process metrics
	// that could not be judged in the Canary phase (e.g. full histograms)
	CanaryProgress_Judging = "judging"
	// Reporting indicates that the operator is trying to report results of the test
	// to the pre-configured destination (e.g. github repository)
	CanaryProgress_Reporting = "reporting"
	// Promithing indicates that the current Krane deployment is being reflected into
	// the stable Krane deployment
	CanaryProgress_Promoting = "promoting"
)

// CanaryStatus defines the observed state of Canary
// +k8s:openapi-gen=true
type CanaryStatus struct {
	Progress CanaryProgress `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Canary is the Schema for the canaries API
// +k8s:openapi-gen=true
type Canary struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CanarySpec   `json:"spec,omitempty"`
	Status CanaryStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CanaryList contains a list of Canary
type CanaryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Canary `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Canary{}, &CanaryList{})
}
