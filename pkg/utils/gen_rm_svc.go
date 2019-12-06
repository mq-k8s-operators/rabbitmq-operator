package utils

import (
	v1 "github.com/lesolise/rabbitmq-operator/pkg/apis/lesolise/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewManagementSvcForCR(cr *v1.RabbitMQ) *corev1.Service {
	managePort := corev1.ServicePort{Port: 15672,Name:"management"}
	ports := make([]corev1.ServicePort, 0)
	ports = append(ports, managePort)

	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rmq-m-svc-" + cr.Name,
			Namespace: cr.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: ports,
			Selector: map[string]string{
				"app": "rmq-node-" + cr.Name,
			},
		},
	}
}