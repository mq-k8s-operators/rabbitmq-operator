package utils

import v1 "github.com/lesolise/rabbitmq-operator/pkg/apis/lesolise/v1"

func CheckCR(cr *v1.RabbitMQ) bool {
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

	if cr.Spec.ProxyImage == "" {
		cr.Spec.ProxyImage = "jianzhiunique/mqproxy:latest"
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

	if cr.Spec.ProxyDiskLimit == "" {
		cr.Spec.ProxyDiskLimit = "10Gi"
		changed = true
	}

	if cr.Spec.ProxyDiskRequest == "" {
		cr.Spec.ProxyDiskRequest = "1Gi"
		changed = true
	}

	if cr.Spec.ProxyMemoryLimit == "" {
		cr.Spec.ProxyMemoryLimit = "2Gi"
		changed = true
	}

	if cr.Spec.ProxyMemoryRequest == "" {
		cr.Spec.ProxyMemoryRequest = "1Gi"
		changed = true
	}

	if cr.Spec.ProxyCpuLimit == "" {
		cr.Spec.ProxyCpuLimit = "2000m"
		changed = true
	}

	if cr.Spec.ProxyCpuRequest == "" {
		cr.Spec.ProxyCpuRequest = "500m"
		changed = true
	}

	if cr.Spec.ManagerHost == "" {
		cr.Spec.ManagerHost = "rmq.cloudmq.com"
		changed = true
	}

	return changed
}
