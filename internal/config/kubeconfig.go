package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// KubeConfig represents the structure of the kubeconfig file
type KubeConfig struct {
	CurrentContext string    `yaml:"current-context"`
	Contexts       []Context `yaml:"contexts"`
}

// FullKubeConfig represents the complete kubeconfig structure
type FullKubeConfig struct {
	APIVersion     string                   `yaml:"apiVersion,omitempty"`
	Kind           string                   `yaml:"kind,omitempty"`
	CurrentContext string                   `yaml:"current-context"`
	Contexts       []Context                `yaml:"contexts"`
	Clusters       []map[string]interface{} `yaml:"clusters,omitempty"`
	Users          []map[string]interface{} `yaml:"users,omitempty"`
	Preferences    map[string]interface{}   `yaml:"preferences,omitempty"`
}

// Context represents a single context in the kubeconfig
type Context struct {
	Name    string        `yaml:"name"`
	Context ContextDetail `yaml:"context"`
}

// ContextDetail contains the cluster, user, and namespace for a context
type ContextDetail struct {
	Cluster   string `yaml:"cluster"`
	User      string `yaml:"user"`
	Namespace string `yaml:"namespace,omitempty"`
}

// Load reads and parses the kubeconfig file
func Load() (*KubeConfig, string, error) {
	path := GetPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}

	var config KubeConfig
	err = yaml.Unmarshal(data, &config)
	return &config, path, err
}

// GetPath returns the kubeconfig path (from KUBECONFIG env or default)
func GetPath() string {
	if envPath := os.Getenv("KUBECONFIG"); envPath != "" {
		return envPath
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".kube", "config")
}

// GetContextNamespaces returns available namespaces for a context
// This could fetch from cluster or allow custom input
func GetContextNamespaces(contextName string) ([]string, error) {
	// Fetch from cluster using kubectl
	cmd := exec.Command("kubectl", "--context", contextName, "get", "namespaces", "-o", "jsonpath={.items[*].metadata.name}")
	output, err := cmd.Output()
	if err != nil {
		// Return common default namespaces if kubectl fails
		return []string{"default", "kube-system", "kube-public", "kube-node-lease"}, nil
	}

	// Parse the output - namespaces are space-separated
	namespacesStr := string(output)
	if namespacesStr == "" {
		return []string{"default"}, nil
	}

	// Split by spaces and filter empty strings
	var namespaces []string
	for _, ns := range strings.Fields(namespacesStr) {
		if ns != "" {
			namespaces = append(namespaces, ns)
		}
	}

	if len(namespaces) == 0 {
		return []string{"default"}, nil
	}

	return namespaces, nil
}

// UpdateContextNamespace updates the namespace for a context
func UpdateContextNamespace(configPath, contextName, namespace string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	// Load FULL config to preserve all fields
	var config FullKubeConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return err
	}

	// Find and update the context
	for i, ctx := range config.Contexts {
		if ctx.Name == contextName {
			config.Contexts[i].Context.Namespace = namespace
			break
		}
	}

	// Write back with proper formatting
	updatedData, err := yaml.Marshal(&config)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, updatedData, 0600)
}
