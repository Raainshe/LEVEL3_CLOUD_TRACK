package kube

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var serviceGVR = schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "services",
}

// GetRedisServicePort returns the external port for the Redis instance's Service.
// Looks up Service rfr-{name}, returns NodePort if type is NodePort, otherwise
// the Service port. Returns 0, nil if Service not found (caller can use env fallback).
func GetRedisServicePort(ctx context.Context, client dynamic.Interface, namespace, name string) (int, error) {
	svcName := "rfr-" + name
	obj, err := client.Resource(serviceGVR).Namespace(namespace).Get(ctx, svcName, metav1.GetOptions{})
	if err != nil {
		return 0, nil
	}

	ports, found, err := unstructured.NestedSlice(obj.Object, "spec", "ports")
	if err != nil || !found || len(ports) == 0 {
		return 0, nil
	}

	portSpec, ok := ports[0].(map[string]interface{})
	if !ok {
		return 0, nil
	}

	if nodePort, found, _ := unstructured.NestedInt64(portSpec, "nodePort"); found && nodePort > 0 {
		return int(nodePort), nil
	}

	if port, found, _ := unstructured.NestedInt64(portSpec, "port"); found && port > 0 {
		return int(port), nil
	}

	return 0, nil
}
