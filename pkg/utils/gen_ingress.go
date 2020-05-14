package utils

import (
	v1 "github.com/lesolise/rabbitmq-operator/pkg/apis/lesolise/v1"
	v1beta12 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func NewIngressForCRIfNotExists(cr *v1.RabbitMQ) *v1beta12.Ingress {
	//通过比较资源的ns与ingress的配置ns来确定servicename要不要用external
	servicename := "rmq-m-svc-" + cr.Namespace + "-" + cr.Name
	if cr.Namespace != cr.Spec.IngressNamespace {
		servicename = "external-" + servicename
	}

	var annotations map[string]string
	annotations = make(map[string]string)
	annotations["nginx.ingress.kubernetes.io/rewrite-target"] = "/$1"

	paths := make([]v1beta12.HTTPIngressPath, 0)
	path := v1beta12.HTTPIngressPath{
		Path: "/rmq-" + cr.Namespace + "-" + cr.Name + "/(.*)",
		Backend: v1beta12.IngressBackend{
			ServicePort: intstr.IntOrString{
				IntVal: 15672,
			},
			ServiceName: servicename,
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
			Name:        "mq-ingress",
			Namespace:   cr.Spec.IngressNamespace,
			Annotations: annotations,
		},
		Spec: v1beta12.IngressSpec{
			Rules: rules,
		},
	}
}

func AppendManagementPathToIngress(cr *v1.RabbitMQ, ingress *v1beta12.Ingress) {
	//通过比较资源的ns与ingress的配置ns来确定servicename要不要用external
	servicename := "rmq-m-svc-" + cr.Namespace + "-" + cr.Name
	if cr.Namespace != cr.Spec.IngressNamespace {
		servicename = "external-" + servicename
	}

	path := v1beta12.HTTPIngressPath{
		Path: "/rmq-" + cr.Namespace + "-" + cr.Name + "/(.*)",
		Backend: v1beta12.IngressBackend{
			ServicePort: intstr.IntOrString{
				IntVal: 15672,
			},
			ServiceName: servicename,
		},
	}

	var found bool

	for _, path := range ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths {
		if path.Path == "/rmq-" + cr.Namespace + "-" + cr.Name + "/(.*)" {
			found = true
			break
		}
	}

	if !found {
		ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths = append(ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths, path)
	}
}

func DeleteManagementPathFromIngress(cr *v1.RabbitMQ, ingress *v1beta12.Ingress) {

	paths := ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths

	for index, path := range ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths {
		if path.Path == "/rmq-" + cr.Namespace + "-" + cr.Name + "/(.*)" {
			ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths = append(paths[:index], paths[index+1:]...)
			break
		}
	}
}

func AppendRabbitMQToolsPathToIngress(cr *v1.RabbitMQ, ingress *v1beta12.Ingress) {
	//通过比较资源的ns与ingress的配置ns来确定servicename要不要用external
	servicename := "rmq-tools-svc-" + cr.Namespace + "-" + cr.Name
	if cr.Namespace != cr.Spec.IngressNamespace {
		servicename = "external-" + servicename
	}

	path := v1beta12.HTTPIngressPath{
		Path: cr.Status.RabbitmqManagerPath + "(.*)",
		Backend: v1beta12.IngressBackend{
			ServicePort: intstr.IntOrString{
				IntVal: 8888,
			},
			ServiceName: servicename,
		},
	}

	var found bool

	for _, path := range ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths {
		if path.Path == cr.Status.RabbitmqManagerPath + "(.*)" {
			found = true
			break
		}
	}

	if !found {
		ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths = append(ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths, path)
	}
}

func DeleteRabbitMQToolsPathFromIngress(cr *v1.RabbitMQ, ingress *v1beta12.Ingress) {

	paths := ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths

	for index, path := range ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths {
		if path.Path == cr.Status.RabbitmqManagerPath + "(.*)" {
			ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths = append(paths[:index], paths[index+1:]...)
			break
		}
	}
}
