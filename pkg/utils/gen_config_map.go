package utils

import (
	v1 "github.com/lesolise/rabbitmq-operator/pkg/apis/lesolise/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewConfigMapForCR(cr *v1.RabbitMQ) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rmq-config-" + cr.Name,
			Namespace: cr.Namespace,
		},
		//TODO rabbitmq_delayed_message_exchange this need a init container
		Data: map[string]string{
			"enabled_plugins": "[rabbitmq_management,rabbitmq_peer_discovery_k8s,rabbitmq_federation_management,rabbitmq_shovel_management,rabbitmq_random_exchange].",
			"rabbitmq.conf": "cluster_formation.peer_discovery_backend  = rabbit_peer_discovery_k8s\n" +
				"cluster_formation.k8s.host = kubernetes.default.svc.cluster.local\n" +
				"cluster_formation.k8s.address_type = hostname\n" +

				"cluster_formation.node_cleanup.interval = 30\n" +
				"cluster_formation.node_cleanup.only_log_warning = false\n" +

				"queue_master_locator=min-masters\n" +
				"cluster_partition_handling = pause_minority\n" +
				"default_pass = " + cr.Status.RabbitmqManagerPassword + "\n" +
				"default_user = " + cr.Status.RabbitmqManagerUsername + "\n" +

				"tcp_listen_options.backlog = 4096\n" +
				"tcp_listen_options.nodelay = true\n" +
				"tcp_listen_options.exit_on_close = false\n" +
				"tcp_listen_options.keepalive = true\n" +
				"tcp_listen_options.send_timeout = 15000\n" +
				"tcp_listen_options.buffer = 196608\n" +
				"tcp_listen_options.sndbuf = 196608\n" +
				"tcp_listen_options.recbuf = 196608\n" +
				"vm_memory_high_watermark.relative = 0.45\n" +
				"vm_memory_high_watermark_paging_ratio = 0.5\n" +
				"disk_free_limit.relative = 1.2\n" +
				"collect_statistics_interval = 30000\n" +
				"log.file.rotation.date = $D0\n",
		},
	}
}
