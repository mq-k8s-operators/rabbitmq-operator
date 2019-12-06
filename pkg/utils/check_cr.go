package utils

import v1 "github.com/lesolise/rabbitmq-operator/pkg/apis/lesolise/v1"

func CheckCR(cr * v1.RabbitMQ) bool{
	var changed bool
	changed = false

	if cr.Spec.Size == 0 || cr.Spec.Size < 3 {
		cr.Spec.Size = 3
		changed = true
	}

	if cr.Spec.Image == "" {
		cr.Spec.Image = "rabbitmq:3.8"
		changed = true
	}

	if cr.Spec.DiskLimit == "" {
		cr.Spec.DiskLimit = "500Gi"
		changed = true
	}

	if cr.Spec.DiskRequest == "" {
		cr.Spec.DiskRequest = "1Gi"
		changed = true
	}

	if cr.Spec.MemoryLimit == "" {
		cr.Spec.MemoryLimit = "16Gi"
		changed = true
	}

	if cr.Spec.MemoryRequest == "" {
		cr.Spec.MemoryRequest = "1Gi"
		changed = true
	}

	if cr.Spec.CpuLimit == "" {
		cr.Spec.CpuLimit = "2000m"
		changed = true
	}

	if cr.Spec.CpuRequest == "" {
		cr.Spec.CpuRequest = "500m"
		changed = true
	}

	if cr.Spec.ManagerHost == "" {
		cr.Spec.ManagerHost = ".rmq.cloudmq.com"
		changed = true
	}

	return changed
}