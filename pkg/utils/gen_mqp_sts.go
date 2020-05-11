package utils

import (
	v1 "github.com/lesolise/rabbitmq-operator/pkg/apis/lesolise/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func NewProxyForCR(cr *v1.RabbitMQ) *appsv1.Deployment {
	var replica int32 = 2

	limit := resource.MustParse(cr.Spec.ProxyDiskLimit)
	pv := make([]corev1.Volume, 0)
	pv = append(pv, corev1.Volume{
		Name: "rmq-mqp-data",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{
				Medium:    "",
				SizeLimit: &limit,
			},
		},
	})

	/*accessModes := make([]corev1.PersistentVolumeAccessMode, 0)
	accessModes = append(accessModes, corev1.ReadWriteOnce)
	pvc := make([]corev1.PersistentVolumeClaim, 0)
	pvc = append(pvc, corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: "rmq-mqp-data",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &cr.Spec.StorageClassName,
			Resources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(cr.Spec.ProxyDiskLimit),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(cr.Spec.ProxyDiskRequest),
				},
			},
			AccessModes: accessModes,
		},
	})*/

	containers := make([]corev1.Container, 0)
	ports := make([]corev1.ContainerPort, 0)
	ports = append(ports, corev1.ContainerPort{
		Name:          "rmq-mqp-port",
		ContainerPort: 8080,
		Protocol:      "TCP",
	})

	envs := make([]corev1.EnvVar, 0)
	envs = append(envs,
		corev1.EnvVar{
			Name:  "rabbit.address.host",
			Value: "rmq-svc-" + cr.Name,
		},
		corev1.EnvVar{
			Name:  "rabbit.address.port",
			Value: "5672",
		},
		corev1.EnvVar{
			Name:  "proxy.config.rabbitmq",
			Value: "true",
		},
		corev1.EnvVar{
			Name:  "proxy.config.kafka",
			Value: "false",
		},
		corev1.EnvVar{
			Name:  "logging.path",
			Value: "/data/mqp",
		},
	)
	vms := make([]corev1.VolumeMount, 0)
	vms = append(vms, corev1.VolumeMount{
		Name:      "rmq-mqp-data",
		MountPath: "/data/mqp",
	})
	healthCheck := corev1.Probe{
		Handler: corev1.Handler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.IntOrString{
					IntVal: 8080,
				},
			},
		},
		InitialDelaySeconds: 15,
		PeriodSeconds:       20,
	}
	c := corev1.Container{
		Name:           "rmq-mqp",
		Image:          cr.Spec.ProxyImage,
		Ports:          ports,
		Env:            envs,
		VolumeMounts:   vms,
		LivenessProbe:  &healthCheck,
		ReadinessProbe: &healthCheck,
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse(cr.Spec.ProxyMemoryLimit),
				corev1.ResourceCPU:    resource.MustParse(cr.Spec.ProxyCpuLimit),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse(cr.Spec.ProxyMemoryRequest),
				corev1.ResourceCPU:    resource.MustParse(cr.Spec.ProxyCpuRequest),
			},
		},
	}
	containers = append(containers, c)

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rmq-mqp-sts-" + cr.Name,
			Namespace: cr.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replica,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "rmq-mqp-" + cr.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":     "rmq-mqp-" + cr.Name,
						"cluster": "rmq-" + cr.Namespace + "-" + cr.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: containers,
					Volumes:    pv,
				},
			},
			Strategy:                appsv1.DeploymentStrategy{},
			MinReadySeconds:         0,
			RevisionHistoryLimit:    nil,
			Paused:                  false,
			ProgressDeadlineSeconds: nil,
		},

		/*
			Spec: appsv1.StatefulSetSpec{
				Replicas:    &replica,
				ServiceName: "rmq-mqp-svc-" + cr.Name,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"app": "rmq-mqp-" + cr.Name},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "rmq-mqp-" + cr.Name,
						},
					},
					Spec: corev1.PodSpec{
						Containers: containers,
					},
				},
				// for data store
				VolumeClaimTemplates: pvc,
			},
			Status: appsv1.StatefulSetStatus{},*/
	}
}
