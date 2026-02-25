package kube

import (
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func loadKubeConfig() (*rest.Config, error) {

	if cfg, err := rest.InClusterConfig(); err == nil {
		return cfg, nil
	}

	if kubeconfig := os.Getenv("KUBECONFIG_PATH"); kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		defaultPath := filepath.Join(home, ".kube", "config")
		if _, statErr := os.Stat(defaultPath); statErr == nil {
			return clientcmd.BuildConfigFromFlags("", defaultPath)
		}
	}

	return nil, fmt.Errorf("Could not load kubeconfig")
}

func NewClient() (dynamic.Interface, error) {
	cfg, err := loadKubeConfig()
	if err != nil {
		return nil, err
	}

	client, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return client, nil
}
