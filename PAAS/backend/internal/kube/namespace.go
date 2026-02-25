package kube

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var namespaceGVR = schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "namespaces",
}

// ListNamespaceNames returns names of all namespaces (for poller fallback when cluster-wide list is forbidden).
func ListNamespaceNames(ctx context.Context, client dynamic.Interface) ([]string, error) {
	list, err := client.Resource(namespaceGVR).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(list.Items))
	for i := range list.Items {
		name, _, _ := unstructured.NestedString(list.Items[i].Object, "metadata", "name")
		if name != "" {
			names = append(names, name)
		}
	}
	return names, nil
}

func EnsureNamespace(ctx context.Context, client dynamic.Interface, name string) error {
	if name == "" || name == "default" || name == "kube-system" || name == "kube-public" {
		return nil
	}

	_, err := client.Resource(namespaceGVR).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		return nil
	}

	ns := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name": name,
			},
		},
	}
	_, err = client.Resource(namespaceGVR).Create(ctx, ns, metav1.CreateOptions{})
	return err
}
