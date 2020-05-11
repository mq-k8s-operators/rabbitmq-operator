package apis

import (
	v12 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "github.com/lesolise/rabbitmq-operator/pkg/apis/lesolise/v1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1.SchemeBuilder.AddToScheme)
	AddToSchemes = append(AddToSchemes, v12.SchemeBuilder.AddToScheme)
}
