package rabbitmq

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rabbitmqv1alpha1 "rabbitmq-operator/pkg/apis/rabbitmq/v1alpha1"
)

func newConfigMap(cr *rabbitmqv1alpha1.Rabbitmq) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rabbitmq-config",
			Namespace: "default",
		},
		Data: cr.Spec.Data,
	}
}

func newPVC(cr *rabbitmqv1alpha1.Rabbitmq) *corev1.PersistentVolumeClaim {
	var scn string
	scn = "managed-nfs-storage"
	return &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rabbitmq-data-claim",
			Namespace: "default",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteMany,
			},
			StorageClassName: &scn,
			Resources: corev1.ResourceRequirements{
				Requests: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceStorage: cr.Spec.Storage,
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"release": "rabbitmq-data"},
			},
		},
	}
}

func newRabbitmqService(cr *rabbitmqv1alpha1.Rabbitmq) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rabbitmq",
			Namespace: "default",
			Labels:    map[string]string{"app": "rabbitmq", "type": "LoadBalancer"},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				corev1.ServicePort{
					Port:     15672,
					Name:     "test1",
					Protocol: corev1.ProtocolTCP,
					NodePort: 32001,
				},
				corev1.ServicePort{
					Port:     5672,
					Name:     "test2",
					Protocol: corev1.ProtocolTCP,
					NodePort: 32002,
				},
			},
			Selector: map[string]string{"app": "rabbitmq"},
		},
	}
}

func newStatefulSet(cr *rabbitmqv1alpha1.Rabbitmq) *appsv1.StatefulSet {

	//set metadata -> label
	alabels := map[string]string{"app": "rabbitmq"}
	//set ImagePullSecret
	//var name corev1.LocalObjectReference
	//name.Name = "regsecret"
	//secrets := []corev1.LocalObjectReference{name}
	//set container
	var container corev1.Container
	container.Name = "rabbitmq"
	container.Image = cr.Spec.Image
	container.ImagePullPolicy = corev1.PullIfNotPresent
	//limits := map[corev1.ResourceName]resource.Quantity{
	//	corev1.ResourceCPU: resource.Quantity{
	//		Format: "256Mi",
	//	},
	//	corev1.ResourceMemory: resource.Quantity{
	//		Format: "150M",
	//	},
	//}
	//requests := map[corev1.ResourceName]resource.Quantity{
	//	corev1.ResourceCPU: resource.Quantity{
	//		Format: "512Mi",
	//	},
	//	corev1.ResourceMemory: resource.Quantity{
	//		Format: "150M",
	//	},
	//}
	//container.Resources.Limits = limits
	//container.Resources.Requests = requests
	container.VolumeMounts = []corev1.VolumeMount{
		//corev1.VolumeMount{
		//	Name:      "rabbitmq-data",
		//	MountPath: "/var/lib/rabbitmq/mnesia",
		//},
		corev1.VolumeMount{
			Name:      "config",
			MountPath: "/etc/rabbitmq",
		},
	}
	container.Env = cr.Spec.Envs
	container.Ports = []corev1.ContainerPort{
		corev1.ContainerPort{
			Name:          "http",
			ContainerPort: 15672,
			Protocol:      "TCP",
		},
		corev1.ContainerPort{
			Name:          "amqp",
			ContainerPort: 5672,
			Protocol:      "TCP",
		},
	}
	container.LivenessProbe = &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{"rabbitmq-diagnostics", "status"},
			},
		},
		InitialDelaySeconds: 60,
		TimeoutSeconds:      15,
		PeriodSeconds:       60,
	}
	container.ReadinessProbe = &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{"rabbitmq-diagnostics", "status"},
			},
		},
		InitialDelaySeconds: 20,
		TimeoutSeconds:      10,
		PeriodSeconds:       60,
	}
	var te int64
	te = 10
	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rabbitmq",
			Namespace: "default",
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: "rabbitmq",
			Replicas:    &cr.Spec.Size,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: alabels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName:            "rabbitmq-operator",
					TerminationGracePeriodSeconds: &te,
					NodeSelector:                  map[string]string{"kubernetes.io/os": "linux"},
					Containers:                    []corev1.Container{container},
					Volumes: []corev1.Volume{
						//corev1.Volume{
						//	Name: "rabbitmq-data",
						//	VolumeSource: corev1.VolumeSource{
						//		PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						//			ClaimName: "rabbitmq-data-claim"},
						//	},
						//},
						corev1.Volume{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: "rabbitmq-config"},
									Items: []corev1.KeyToPath{
										corev1.KeyToPath{
											Key:  "rabbitmq.conf",
											Path: "rabbitmq.conf",
										},
										corev1.KeyToPath{
											Key:  "enabled_plugins",
											Path: "enabled_plugins",
										},
									},
								},
							},
						},
					},
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: alabels,
			},
		},
	}
}