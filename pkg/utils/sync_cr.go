package utils

import (
	appsv1 "k8s.io/api/apps/v1"
)

func SyncRabbitMQSts(old *appsv1.StatefulSet, new *appsv1.StatefulSet) {
	old.Spec.Replicas = new.Spec.Replicas
	var newConfig = make(map[string]string)
	for _, newEnv := range new.Spec.Template.Spec.Containers[0].Env {
		newConfig[newEnv.Name] = newEnv.Value
	}
	for key, env := range old.Spec.Template.Spec.Containers[0].Env {
		old.Spec.Template.Spec.Containers[0].Env[key].Value = newConfig[env.Name]
	}
}
