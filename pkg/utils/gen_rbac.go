package utils

import (
	v1 "github.com/lesolise/rabbitmq-operator/pkg/apis/lesolise/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/rbac/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewServiceAccountForCR(cr *v1.RabbitMQ) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rabbitmq",
			Namespace: cr.Namespace,
		},
	}
}

func NewRoleForCR(cr *v1.RabbitMQ) *v1beta1.Role {
	apiGroups := make([]string, 0)
	apiGroups = append(apiGroups, "")
	resources := make([]string, 0)
	resources = append(resources, "endpoints")
	verbs := make([]string, 0)
	verbs = append(verbs, "get")

	roles := make([]v1beta1.PolicyRule, 0)
	role := v1beta1.PolicyRule{
		APIGroups: apiGroups,
		Resources: resources,
		Verbs:     verbs,
	}
	roles = append(roles, role)

	return &v1beta1.Role{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "rbac.authorization.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rabbitmq-peer-discovery-rbac",
			Namespace: cr.Namespace,
		},
		Rules: roles,
	}
}

func NewRoleBindingForCR(cr *v1.RabbitMQ) *v1beta1.RoleBinding {
	subjects := make([]v1beta1.Subject, 0)
	subject := v1beta1.Subject{
		Kind: "ServiceAccount",
		Name: "rabbitmq",
	}
	subjects = append(subjects, subject)

	return &v1beta1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1beta1",
			Kind:       "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rabbitmq-peer-discovery-rbac",
			Namespace: cr.Namespace,
		},
		Subjects: subjects,
		RoleRef: v1beta1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "rabbitmq-peer-discovery-rbac",
		},
	}
}
