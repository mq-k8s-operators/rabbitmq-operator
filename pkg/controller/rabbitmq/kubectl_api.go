package rabbitmq

import (
	corev1 "k8s.io/api/core/v1"
	rbac1 "k8s.io/api/rbac/v1"
	storage1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rabbitmqv1alpha1 "rabbitmq-operator/pkg/apis/rabbitmq/v1alpha1"
)

//根据需求选择是否需要设置
func newService(cr *rabbitmqv1alpha1.Rabbitmq) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rabbitmq",
			Labels:    map[string]string{"app": "rabbitmq"},
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None",
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Port: 5672,
					Name: "amqp",
				},
			},
			Selector: map[string]string{"app": "rabbitmq"},
		},
	}
}

func newServiceAccount(cr *rabbitmqv1alpha1.Rabbitmq) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rabbitmq",
			Namespace: "default",
		},
	}
}

func newPV(cr *rabbitmqv1alpha1.Rabbitmq) *corev1.PersistentVolume {
	return &corev1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolume",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "rabbitmq-data",
			Labels: map[string]string{
				"release": "rabbitmq-data",
			},
			Namespace: "default",
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceStorage: resource.Quantity{
					Format: "2Gi",
				},
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteMany,
			},
			StorageClassName: "managed-nfs-storage",
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				NFS: &corev1.NFSVolumeSource{
					Server: "/home/k8s/nfs/data/pv001",
					Path:   "127.0.0.1",
				},
			},
		},
	}
}

func newStorageClass(cr *rabbitmqv1alpha1.Rabbitmq) *storage1.StorageClass {

	return &storage1.StorageClass{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StorageClass",
			APIVersion: "storage.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "managed-nfs-storage",
			Namespace: "default",
		},
		Provisioner: "fuseim.pri/ifs",
		Parameters:  map[string]string{"archiveOnDelete": "false"},
	}
}

func newRole(cr *rabbitmqv1alpha1.Rabbitmq) *rbac1.Role {
	return &rbac1.Role{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Role",
			APIVersion: "rbac.authorization.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "endpoint-reader",
			Namespace: "default",
		},
		Rules: []rbac1.PolicyRule{
			rbac1.PolicyRule{
				Verbs:     []string{"get"},
				APIGroups: []string{""},
				Resources: []string{"endpoints"},
			},
		},
	}
}

func newRoleBing(cr *rabbitmqv1alpha1.Rabbitmq) *rbac1.RoleBinding {
	return &rbac1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "endpoint-reader",
		},
		Subjects: []rbac1.Subject{
			rbac1.Subject{
				Kind: "ServiceAccount",
				Name: "rabbitmq",
			},
		},
		RoleRef: rbac1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "endpoint-reader",
		},
	}
}
