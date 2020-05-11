package utils

import (
	v12 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "github.com/lesolise/rabbitmq-operator/pkg/apis/lesolise/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewSvcMonitorForCR(cr *v1.RabbitMQ) *v12.ServiceMonitor {
	name := "rmq-metrics-" + cr.Name

	endpoint := v12.Endpoint{
		Port:        cr.Status.RabbitmqUrl + "-metrics-port",
		Interval:    "10s",
		HonorLabels: false,
	}
	endpoints := make([]v12.Endpoint, 0)
	endpoints = append(endpoints, endpoint)

	return &v12.ServiceMonitor{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "monitoring.coreos.com/v1",
			Kind:       "ServiceMonitor",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: cr.Namespace,
			Labels: map[string]string{
				"app":      name,
				"release":  "prometheus-operator",
				"heritage": "Tiller",
				"k8s-app":  "rmq-metrics",
			},
		},
		Spec: v12.ServiceMonitorSpec{
			JobLabel:  name + "-" + cr.Namespace,
			Endpoints: endpoints,
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"mq-metrics": cr.Status.RabbitmqUrl + "-metrics",
				},
			},
			NamespaceSelector: v12.NamespaceSelector{
				MatchNames: []string{cr.Namespace},
			},
		},
	}
}
