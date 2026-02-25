package models

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type RedisInstance struct {
	ID               string    `json:"id" bson:"_id"`
	Name             string    `json:"name" bson:"name"`
	Namespace        string    `json:"namespace" bson:"namespace"`
	RedisReplicas    int       `json:"redisReplicas" bson:"redis_replicas"`
	SentinelReplicas int       `json:"sentinelReplicas" bson:"sentinel_replicas"`
	Status           string    `json:"status" bson:"status"`
	CreatedAt        time.Time `json:"createdAt" bson:"created_at"`
	UpdatedAt        time.Time `json:"updatedAt" bson:"updated_at"`

	ExternalHost string `json:"externalHost,omitempty" bson:"-"`
	ExternalPort int    `json:"externalPort,omitempty" bson:"-"`
	RedisCLI     string `json:"redisCli,omitempty" bson:"-"`
}

type CreateInstanceRequest struct {
	Name             string `json:"name" bson:"name"`
	Namespace        string `json:"namespace" bson:"namespace"`
	RedisReplicas    int    `json:"redisReplicas" bson:"redis_replicas"`
	SentinelReplicas int    `json:"sentinelReplicas" bson:"sentinel_replicas"`
}

type DeleteInstanceRequest struct {
	Namespace string `json:"namespace" bson:"namespace"`
}

type UpdateInstanceRequest struct {
	Namespace        *string `json:"namespace,omitempty" bson:"namespace,omitempty"`
	RedisReplicas    *int    `json:"redisReplicas,omitempty" bson:"redis_replicas,omitempty"`
	SentinelReplicas *int    `json:"sentinelReplicas,omitempty" bson:"sentinel_replicas,omitempty"`
}

func (r *RedisInstance) GetConnectionInfo(portOverride int) error {
	host := os.Getenv("REDIS_GATEWAY_HOST")
	if host == "" {
		return errors.New("REDIS_GATEWAY_HOST is not set")
	}

	port := portOverride
	if port <= 0 {
		port = 6379
		if portStr := os.Getenv("REDIS_GATEWAY_PORT"); portStr != "" {
			if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
				port = p
			} else {
				return errors.New("REDIS_GATEWAY_PORT is not a valid integer")
			}
		}
	}

	r.ExternalHost = host
	r.ExternalPort = port
	r.RedisCLI = fmt.Sprintf("redis-cli -h %s -p %d", host, port)
	return nil
}

func (r *RedisInstance) ConvertUnstructuredToRedisInstace(item *unstructured.Unstructured) {
	r.ID = item.GetName()
	r.Name = item.GetName()
	r.Namespace = item.GetNamespace()

	r.RedisReplicas = 0
	r.SentinelReplicas = 0
	r.Status = "-"
	r.CreatedAt = item.GetCreationTimestamp().Time
	r.UpdatedAt = item.GetCreationTimestamp().Time

	spec, found, _ := unstructured.NestedMap(item.Object, "spec")
	if found {
		if redis, ok := spec["redis"].(map[string]interface{}); ok {
			if replicas, ok := redis["replicas"].(int64); ok {
				r.RedisReplicas = int(replicas)
			}
		}
		if sentinel, ok := spec["sentinel"].(map[string]interface{}); ok {
			if replicas, ok := sentinel["replicas"].(int64); ok {
				r.SentinelReplicas = int(replicas)
			}
		}
	}

	r.Status = extractStatusFromUnstructured(item)
}

func extractStatusFromUnstructured(item *unstructured.Unstructured) string {
	_, hasStatus, _ := unstructured.NestedMap(item.Object, "status")
	if !hasStatus {
		return "Unknown"
	}
	if phase, found, _ := unstructured.NestedString(item.Object, "status", "phase"); found && phase != "" {
		return normalizeStatus(phase)
	}
	if state, found, _ := unstructured.NestedString(item.Object, "status", "state"); found && state != "" {
		return normalizeStatus(state)
	}
	if st, found, _ := unstructured.NestedString(item.Object, "status", "status"); found && st != "" {
		return normalizeStatus(st)
	}
	conditions, found, _ := unstructured.NestedSlice(item.Object, "status", "conditions")
	if found && len(conditions) > 0 {
		if s := extractStatusFromConditions(conditions); s != "" {
			return s
		}
	}

	return "Unknown"
}

func extractStatusFromConditions(conditions []interface{}) string {
	priorityTypes := []string{"Ready", "Available", "Reconciling", "Progressing"}

	for _, pt := range priorityTypes {
		for _, c := range conditions {
			cond, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			condType, _ := cond["type"].(string)
			if condType != pt {
				continue
			}
			status, _ := cond["status"].(string)
			reason, _ := cond["reason"].(string)

			switch status {
			case "True":
				if pt == "Ready" || pt == "Available" {
					return "Running"
				}
				return normalizeStatus(condType)
			case "False":
				if reason != "" {
					return normalizeStatus(reason)
				}
				if pt == "Ready" || pt == "Available" {
					return "Pending"
				}
				return normalizeStatus(condType)
			}
		}
	}

	if len(conditions) > 0 {
		if cond, ok := conditions[0].(map[string]interface{}); ok {
			if reason, ok := cond["reason"].(string); ok && reason != "" {
				return normalizeStatus(reason)
			}
			if ct, ok := cond["type"].(string); ok && ct != "" {
				return normalizeStatus(ct)
			}
		}
	}

	return ""
}

func normalizeStatus(s string) string {
	switch s {
	case "Running", "Ready", "Available", "Healthy":
		return "Running"
	case "Provisioning", "Pending", "Reconciling", "Progressing", "Creating":
		return "Provisioning"
	case "Error", "Failed", "Degraded":
		return "Failed"
	default:
		return s
	}
}
