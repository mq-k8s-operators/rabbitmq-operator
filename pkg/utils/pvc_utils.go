package utils

import (
	"strconv"
	"strings"
)

const (
	KafkaFinalizer = "cleanUpRabbitMQPVC"
)

func ContainsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

func RemoveString(slice []string, str string) (result []string) {
	for _, item := range slice {
		if item == str {
			continue
		}
		result = append(result, item)
	}
	return result
}

func IsPVCOrphan(zkPvcName string, replicas int32) bool {
	index := strings.LastIndexAny(zkPvcName, "-")
	if index == -1 {
		return false
	}

	ordinal, err := strconv.Atoi(zkPvcName[index+1:])
	if err != nil {
		return false
	}

	return int32(ordinal) >= replicas
}
