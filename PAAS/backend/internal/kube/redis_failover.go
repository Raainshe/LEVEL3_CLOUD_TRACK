package kube

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var RedisFailOver = schema.GroupVersionResource{
	Group:    "databases.spotahome.com",
	Version:  "v1",
	Resource: "redisfailovers",
}

func BuildRedisFailover(name, namespace string, redisReplicas, sentinelReplicas int) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "databases.spotahome.com/v1",
			"kind":       "RedisFailover",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"redis": map[string]interface{}{
					"replicas": int64(redisReplicas),
				},
				"sentinel": map[string]interface{}{
					"replicas": int64(sentinelReplicas),
				},
			},
		},
	}
}
