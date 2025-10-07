package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// ANSI color codes
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Red       = "\033[31m"
	Green     = "\033[32m"
	Yellow    = "\033[33m"
	Blue      = "\033[34m"
	Cyan      = "\033[36m"
	Gray      = "\033[90m"
	BoldCyan  = "\033[1;36m"
	BoldGreen = "\033[1;32m"
)

// KubeConfig represents the structure of the kubeconfig file
type KubeConfig struct {
	CurrentContext string    `yaml:"current-context"`
	Contexts       []Context `yaml:"contexts"`
}

type Context struct {
	Name    string        `yaml:"name"`
	Context ContextDetail `yaml:"context"`
}

type ContextDetail struct {
	Cluster   string `yaml:"cluster"`
	User      string `yaml:"user"`
	Namespace string `yaml:"namespace,omitempty"`
}

func main() {
	// Get kubeconfig path
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %v\n", err)
		os.Exit(1)
	}

	kubeconfigPath := filepath.Join(home, ".kube", "config")

	// Check if KUBECONFIG env var is set
	if envPath := os.Getenv("KUBECONFIG"); envPath != "" {
		kubeconfigPath = envPath
	}

	// Read kubeconfig file
	data, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		fmt.Printf("Error reading kubeconfig file: %v\n", err)
		os.Exit(1)
	}

	// Parse YAML
	var config KubeConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		fmt.Printf("Error parsing kubeconfig: %v\n", err)
		os.Exit(1)
	}

	if len(config.Contexts) == 0 {
		fmt.Println("No contexts found in kubeconfig")
		os.Exit(1)
	}

	// Display contexts
	fmt.Println()
	fmt.Printf("%s╭─────────────────────────────────────────────────────────╮%s\n", BoldCyan, Reset)
	fmt.Printf("%s│              Kubernetes Context Switcher               │%s\n", BoldCyan, Reset)
	fmt.Printf("%s╰─────────────────────────────────────────────────────────╯%s\n", BoldCyan, Reset)
	fmt.Println()

	for i, ctx := range config.Contexts {
		marker := " "
		if ctx.Name == config.CurrentContext {
			marker = BoldGreen + "●" + Reset
			fmt.Printf(" %s [%s%d%s] %s%s%s\n",
				marker,
				Yellow, i+1, Reset,
				BoldGreen, ctx.Name, Reset)
		} else {
			fmt.Printf(" %s [%s%d%s] %s\n",
				marker,
				Yellow, i+1, Reset,
				ctx.Name)
		}
		fmt.Printf("%s      ├─ cluster: %s%s\n", Gray, ctx.Context.Cluster, Reset)
		fmt.Printf("%s      └─ user: %s%s\n", Gray, ctx.Context.User, Reset)
		if i < len(config.Contexts)-1 {
			fmt.Println()
		}
	}

	// Get user choice
	fmt.Println()
	fmt.Printf("%s─────────────────────────────────────────────────────────%s\n", BoldCyan, Reset)
	fmt.Printf("%sSelect context [1-%d] (0 to cancel): %s", Blue, len(config.Contexts), Reset)
	var choice string
	fmt.Scanln(&choice)

	choiceNum, err := strconv.Atoi(choice)
	if err != nil || choiceNum < 0 || choiceNum > len(config.Contexts) {
		fmt.Printf("%sInvalid choice%s\n", Red, Reset)
		os.Exit(1)
	}

	if choiceNum == 0 {
		fmt.Printf("\n%s❌ Cancelled%s\n", Red, Reset)
		os.Exit(0)
	}

	// Update current context
	selectedContext := config.Contexts[choiceNum-1].Name

	// Update only the current-context line in the file
	err = updateCurrentContext(kubeconfigPath, selectedContext)
	if err != nil {
		fmt.Printf("Error updating kubeconfig: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n%s✨ Successfully switched to context: %s%s\n\n", BoldGreen, selectedContext, Reset)
}

// updateCurrentContext updates only the current-context line in the kubeconfig file
func updateCurrentContext(kubeconfigPath, newContext string) error {
	// Read the file
	file, err := os.Open(kubeconfigPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	updated := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Check if this is the current-context line
		if strings.HasPrefix(trimmed, "current-context:") {
			// Preserve the original indentation
			indent := line[:len(line)-len(trimmed)]
			lines = append(lines, fmt.Sprintf("%scurrent-context: %s", indent, newContext))
			updated = true
		} else {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if !updated {
		return fmt.Errorf("current-context line not found in kubeconfig")
	}

	// Write the file back
	output := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(kubeconfigPath, []byte(output), 0644)
}
