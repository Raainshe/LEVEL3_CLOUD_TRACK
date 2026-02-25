package kube

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var (
	statefulSetGVR = schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "statefulsets",
	}
	deploymentGVR = schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}
)

func GetStatusFromStatefulSets(ctx context.Context, client dynamic.Interface, namespace, name string, expectedRedis, expectedSentinel int) string {
	redisName := "rfr-" + name
	sentinelName := "rfs-" + name

	redisStatus := getStatefulSetStatus(ctx, client, namespace, redisName, expectedRedis)
	sentinelStatus := getDeploymentStatus(ctx, client, namespace, sentinelName, expectedSentinel)

	if redisStatus == "" || sentinelStatus == "" {
		return "Provisioning"
	}
	if redisStatus == "Failed" || sentinelStatus == "Failed" {
		return "Failed"
	}
	if redisStatus == "Running" && sentinelStatus == "Running" {
		return "Running"
	}
	return "Provisioning"
}

func getStatefulSetStatus(ctx context.Context, client dynamic.Interface, namespace, name string, expectedReplicas int) string {
	obj, err := client.Resource(statefulSetGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return ""
	}

	replicas, _, _ := unstructured.NestedInt64(obj.Object, "status", "replicas")
	readyReplicas, _, _ := unstructured.NestedInt64(obj.Object, "status", "readyReplicas")

	exp := int64(expectedReplicas)
	if expectedReplicas <= 0 {
		exp = replicas
	}

	if replicas == 0 {
		return "Provisioning"
	}
	if readyReplicas >= exp {
		return "Running"
	}
	return "Provisioning"
}

func getDeploymentStatus(ctx context.Context, client dynamic.Interface, namespace, name string, expectedReplicas int) string {
	obj, err := client.Resource(deploymentGVR).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return ""
	}

	replicas, _, _ := unstructured.NestedInt64(obj.Object, "status", "replicas")
	readyReplicas, _, _ := unstructured.NestedInt64(obj.Object, "status", "readyReplicas")

	exp := int64(expectedReplicas)
	if expectedReplicas <= 0 {
		exp = replicas
	}

	if replicas == 0 {
		return "Provisioning"
	}
	if readyReplicas >= exp {
		return "Running"
	}
	return "Provisioning"
}
