package utils

import (
	v1 "github.com/lesolise/rabbitmq-operator/pkg/apis/lesolise/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewLBSvcForCR(cr *v1.RabbitMQ) *corev1.Service {
	//A headless service must be used to control network identity of the pods (their hostnames), which in turn affect RabbitMQ node names.
	//see: https://www.rabbitmq.com/cluster-formation.html#peer-discovery-k8s
	//In fact, if service is not headless, the network identity of the pods can be resolve by specify ServiceName field on statefulset
	amqpPort := corev1.ServicePort{Port: 5672, Name: "amqp"}
	ports := make([]corev1.ServicePort, 0)
	ports = append(ports, amqpPort)
	cr.Status.RabbitmqUrl = "rmq-svc-" + cr.Name
	cr.Status.RabbitmqPort = "5672"
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Status.RabbitmqUrl,
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
