package apis

import (
	"github.com/petomalina/krane/pkg/apis/krane/v1"
	"github.com/petomalina/krane/pkg/apis/networking/v1alpha3"
	"k8s.io/apimachinery/pkg/runtime"
)

// AddToSchemes may be used to add all resources defined in the project to a Scheme
var AddToSchemes = runtime.SchemeBuilder{
	v1.SchemeBuilder.AddToScheme,
	v1alpha3.SchemeBuilder.AddToScheme,
}

// AddToScheme adds all Resources to the Scheme
func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}
