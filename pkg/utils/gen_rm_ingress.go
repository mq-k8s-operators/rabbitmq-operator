package utils

import (
	v1 "github.com/lesolise/rabbitmq-operator/pkg/apis/lesolise/v1"
	v1beta12 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func NewRabbitMQManagementIngressForCR(cr *v1.RabbitMQ) *v1beta12.Ingress {
	pathStr := "/rmq-" + cr.Namespace + "-" + cr.Name + "/"

	paths := make([]v1beta12.HTTPIngressPath, 0)
	path := v1beta12.HTTPIngressPath{
		Path: pathStr + "(.*)",
		Backend: v1beta12.IngressBackend{
			ServicePort: intstr.IntOrString{
				IntVal: 15672,
			},
			ServiceName: "rmq-m-svc-" + cr.Name,
		},
	}
	paths = append(paths, path)

	rules := make([]v1beta12.IngressRule, 0)
	rule := v1beta12.IngressRule{
		Host: cr.Spec.ManagerHost,
		IngressRuleValue: v1beta12.IngressRuleValue{
			HTTP: &v1beta12.HTTPIngressRuleValue{
				Paths: paths,
			},
		},
	}
	rules = append(rules, rule)

	return &v1beta12.Ingress{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "networking.k8s.io/v1beta1",
			Kind:       "Ingress",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rmq-m-ingress-" + cr.Name,
			Namespace: cr.Namespace,
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/rewrite-target": "/$1",
			},
		},
		Spec: v1beta12.IngressSpec{
			Rules: rules,
		},
	}
}
