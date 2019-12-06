package utils

import (
	appsv1 "k8s.io/api/apps/v1"
)

func SyncRabbitMQSts(old *appsv1.StatefulSet, new *appsv1.StatefulSet) {
	old.Spec.Replicas = new.Spec.Replicas
}
