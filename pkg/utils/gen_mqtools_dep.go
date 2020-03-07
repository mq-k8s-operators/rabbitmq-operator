package utils

import (
	v1 "github.com/lesolise/rabbitmq-operator/pkg/apis/lesolise/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func NewToolsForCR(cr *v1.RabbitMQ) *appsv1.Deployment {
	var replica int32 = 1

	limit := resource.MustParse(cr.Spec.ToolsDiskLimit)
	//this section should use persistent pv, such as ceph, to do
	pv := make([]corev1.Volume, 0)
	pv = append(pv, corev1.Volume{
		Name: "rmq-tools-data",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{
				Medium:    "",
				SizeLimit: &limit,
			},
		},
	})

	containers := make([]corev1.Container, 0)
	ports := make([]corev1.ContainerPort, 0)
	ports = append(ports, corev1.ContainerPort{
		Name:          "rmq-tools-port",
		ContainerPort: 8888,
		Protocol:      "TCP",
	})

	envs := make([]corev1.EnvVar, 0)
	envs = append(envs,
		corev1.EnvVar{
			Name:  "rabbitmq.config.api",
			Value: "rmq-m-svc-" + cr.Name + ":15672",
		},
		corev1.EnvVar{
			Name:  "rabbitmq.config.username",
			Value: cr.Status.RabbitmqManagerUsername,
		},
		corev1.EnvVar{
			Name:  "rabbitmq.config.password",
			Value: cr.Status.RabbitmqManagerPassword,
		},
		corev1.EnvVar{
			Name:  "server.port",
			Value: "8888",
		},
		corev1.EnvVar{
			Name:  "rabbitmq.config.name",
			Value: cr.Namespace + "-" + cr.Name,
		},
		corev1.EnvVar{
			Name:  "rabbitmq.config.admin_ding_url",
			Value: cr.Spec.ToolsAdminDingUrl,
		},
		corev1.EnvVar{
			Name:  "rabbitmq.config.localdb",
			Value: "/data/rabbitmq-localdb-",
		},
	)
	vms := make([]corev1.VolumeMount, 0)
	vms = append(vms, corev1.VolumeMount{
		Name:      "rmq-tools-data",
		MountPath: "/data",
	})
	healthCheck := corev1.Probe{
		Handler: corev1.Handler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.IntOrString{
					IntVal: 8888,
				},
			},
		},
		InitialDelaySeconds: 15,
		PeriodSeconds:       20,
	}
	c := corev1.Container{
		Name:           "rmq-tools",
		Image:          cr.Spec.ToolsImage,
		Ports:          ports,
		Env:            envs,
		VolumeMounts:   vms,
		LivenessProbe:  &healthCheck,
		ReadinessProbe: &healthCheck,
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse(cr.Spec.ToolsMemoryLimit),
				corev1.ResourceCPU:    resource.MustParse(cr.Spec.ToolsCpuLimit),
			},
			Requests: corev1.ResourceList{
				corev1.ResourceMemory: resource.MustParse(cr.Spec.ToolsMemoryRequest),
				corev1.ResourceCPU:    resource.MustParse(cr.Spec.ToolsCpuRequest),
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
			Name:      "rmq-tools-sts-" + cr.Name,
			Namespace: cr.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replica,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "rmq-tools-" + cr.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "rmq-tools-" + cr.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers:         containers,
					Volumes:            pv,
					ServiceAccountName: "rabbitmq-operator",
				},
			},
			Strategy:                appsv1.DeploymentStrategy{},
			MinReadySeconds:         0,
			RevisionHistoryLimit:    nil,
			Paused:                  false,
			ProgressDeadlineSeconds: nil,
		},
	}
}
