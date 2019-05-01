package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CanarySpecDeployments struct {
	Canary   string `json:"canary,omitempty"`
	Baseline string `json:"baseline,omitempty"`
	Base     string `json:"base,omitempty"`
}

// CanarySpec defines the desired state of Canary
// +k8s:openapi-gen=true
type CanarySpec struct {
	Policy string `json:"policy,omitempty"`

	Deployments CanarySpecDeployments `json:"deployments,omitempty"`
}

// CanaryProgress is a progress of the Krane object during its deployment
type CanaryProgress string

const (
	// Initializing indicates that deployment creation and pod initialization
	// is in progress and the operator is waiting until pretest can start
	CanaryProgress_Initializing CanaryProgress = "initializing"
	// Test indicates that the service is being tested using the pretest
	// container to make sure the container can be routed
	CanaryProgress_Testing = "testing"
	// Canary indicates that the operator has split the traffic, created an
	// additional replica of the stable Krane deployment and started the collection
	// of prometheus metrics
	CanaryProgress_Canary = "canary"
	// Reporting indicates that the operator is trying to report results of the test
	// to the pre-configured destination (e.g. github repository)
	CanaryProgress_Reporting = "reporting"
	// Cleanup progress indicates that the canary is now being cleaned up with all
	// created pods and resources
	CanaryProgress_Cleanup = "cleanup"
)

type CanaryPhaseStatus string

const (
	CanaryPhaseStatus_Queued     CanaryPhaseStatus = "queued"
	CanaryPhaseStatus_InProgress                   = "in_progress"
	CanaryPhaseStatus_Success                      = "success"
	CanaryPhaseStatus_Failure                      = "failure"
)

type CanaryConfigPhase struct {
	PodName string            `json:"podName,omitempty"`
	Status  CanaryPhaseStatus `json:"status,omitempty"`
	Message string            `json:"message,omitempty"`
}

// CanaryStatus defines the observed state of Canary
// +k8s:openapi-gen=true
type CanaryStatus struct {
	Progress CanaryProgress `json:"progress,omitempty"`

	Initialization CanaryConfigPhase `json:"initialization,omitempty"`
	Testing        CanaryConfigPhase `json:"testing,omitempty"`
	Canary         CanaryConfigPhase `json:"canary,omitempty"`
	Judging        CanaryConfigPhase `json:"judging,omitempty"`
	Reporting      CanaryConfigPhase `json:"reporting,omitempty"`
	Cleanup        CanaryConfigPhase `json:"cleanup,omitempty"`
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
