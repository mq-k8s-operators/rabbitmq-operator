package utils

import (
	v1 "github.com/lesolise/rabbitmq-operator/pkg/apis/lesolise/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewStsForCR(cr *v1.RabbitMQ) *appsv1.StatefulSet {
	svcName := "rmq-svc-" + cr.Name
	accessModes := make([]corev1.PersistentVolumeAccessMode, 0)
	accessModes = append(accessModes, corev1.ReadWriteOnce)
	pvc := make([]corev1.PersistentVolumeClaim, 0)
	pvc = append(pvc, corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: "rmq-data",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &cr.Spec.StorageClassName,
			Resources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(cr.Spec.DiskLimit),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(cr.Spec.DiskRequest),
				},
			},
			AccessModes: accessModes,
		},
	})

	containers := make([]corev1.Container, 0)
	ports := make([]corev1.ContainerPort, 0)
	ports = append(ports, corev1.ContainerPort{
		Name:          "amqp-port",
		ContainerPort: 5672,
		Protocol:      "TCP",
	})
	ports = append(ports, corev1.ContainerPort{
		Name:          "manage-port",
		ContainerPort: 15672,
		Protocol:      "TCP",
	})

	envs := make([]corev1.EnvVar, 0)
	envs = append(envs,
		corev1.EnvVar{
			Name: "MY_POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		corev1.EnvVar{
			Name: "MY_POD_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		corev1.EnvVar{
			Name:  "K8S_SERVICE_NAME",
			Value: svcName,
		},
		corev1.EnvVar{
			Name:  "RABBITMQ_USE_LONGNAME",
			Value: "true",
		},
		corev1.EnvVar{
			Name:  "RABBITMQ_ERLANG_COOKIE",
			Value: "cookie-" + cr.Name,
		},
		corev1.EnvVar{
			Name:  "RABBITMQ_NODENAME",
			Value: "rabbit@$(MY_POD_NAME).$(K8S_SERVICE_NAME).$(MY_POD_NAMESPACE).svc.cluster.local",
		},
		corev1.EnvVar{
			Name:  "K8S_HOSTNAME_SUFFIX",
			Value: ".$(K8S_SERVICE_NAME).$(MY_POD_NAMESPACE).svc.cluster.local",
		},
	)

	vms := make([]corev1.VolumeMount, 0)
	vms = append(vms, corev1.VolumeMount{
		Name:      "rmq-data",
		MountPath: "/var/lib/rabbitmq/mnesia",
	},
		corev1.VolumeMount{
			Name:      "rmq-config",
			MountPath: "/etc/rabbitmq",
		},
	)
	livenessProbe := &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{"rabbitmq-diagnostics", "status"},
			},
		},
		InitialDelaySeconds: 60,
		TimeoutSeconds:      15,
		PeriodSeconds:       60,
	}
	readinessProbe := &corev1.Probe{
		Handler: corev1.Handler{
			Exec: &corev1.ExecAction{
				Command: []string{"rabbitmq-diagnostics", "status"},
			},
		},
		InitialDelaySeconds: 20,
		TimeoutSeconds:      10,
		PeriodSeconds:       60,
	}

	rmq := corev1.Container{
		Name:           "rmq",
		Image:          cr.Spec.Image,
		Ports:          ports,
		Env:            envs,
		VolumeMounts:   vms,
		LivenessProbe:  livenessProbe,
		ReadinessProbe: readinessProbe,
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse(cr.Spec.MemoryLimit),
				corev1.ResourceCPU:    resource.MustParse(cr.Spec.CpuLimit),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse(cr.Spec.MemoryRequest),
				corev1.ResourceCPU:    resource.MustParse(cr.Spec.CpuRequest),
			},
		},
	}
	containers = append(containers, rmq)

	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rmq-sts-" + cr.Name,
			Namespace: cr.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas:    &cr.Spec.Size,
			ServiceName: svcName,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "rmq-node-" + cr.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "rmq-node-" + cr.Name,
					},
				},
				Spec: corev1.PodSpec{
					// we must specify the service account name for pod
					// see: https://github.com/rabbitmq/rabbitmq-peer-discovery-k8s/pull/16/commits/e0222510e5f69b98c21b92fa14db7ee57887daf6
					ServiceAccountName: "rabbitmq-operator",
					Containers:         containers,
					// for config files generate from config map
					Volumes: []corev1.Volume{
						corev1.Volume{
							Name: "rmq-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: "rmq-config-" + cr.Name},
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
			// for rabbitmq data store
			VolumeClaimTemplates: pvc,
		},
		Status: appsv1.StatefulSetStatus{},
	}
}
