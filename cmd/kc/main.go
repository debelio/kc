package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"kc/internal/config"
	"kc/internal/ui"
)

var (
	contextArg   string
	namespaceArg string
	versionFlag  bool
	version      = "dev" // Set by build flags
)

func init() {
	flag.StringVar(&contextArg, "c", "", "Context name to switch to")
	flag.StringVar(&namespaceArg, "n", "", "Namespace to set (requires -c)")
	flag.BoolVar(&versionFlag, "v", false, "Show version information")
	flag.BoolVar(&versionFlag, "version", false, "Show version information")
	flag.Usage = printUsage
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] [CONTEXT] [NAMESPACE]\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Kubernetes Context and Namespace Switcher\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nExamples:\n")
	fmt.Fprintf(os.Stderr, "  %s                          # Interactive mode\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s prod-cluster             # Switch to context\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s prod-cluster default     # Switch context and set namespace\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s -c prod-cluster          # Switch to context (flag style)\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s -c prod-cluster -n dev   # Switch context and set namespace (flag style)\n", os.Args[0])
}

func main() {
	flag.Parse()

	// Show version and exit
	if versionFlag {
		fmt.Printf("kc version %s\n", version)
		return
	}

	cfg, configPath, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%sError reading kubeconfig: %v%s\n", ui.Red, err, ui.Reset)
		os.Exit(1)
	}

	if len(cfg.Contexts) == 0 {
		fmt.Fprintf(os.Stderr, "%sNo contexts found in kubeconfig%s\n", ui.Red, ui.Reset)
		os.Exit(1)
	}

	// Check for positional arguments
	args := flag.Args()
	if len(args) > 0 && contextArg == "" {
		contextArg = args[0]
	}
	if len(args) > 1 && namespaceArg == "" {
		namespaceArg = args[1]
	}

	// Non-interactive mode: context provided as argument
	if contextArg != "" {
		handleNonInteractiveMode(cfg, configPath, contextArg, namespaceArg)
		return
	}

	// Interactive mode
	handleInteractiveMode(cfg, configPath)
}

func handleNonInteractiveMode(cfg *config.KubeConfig, configPath, contextName, namespace string) {
	// Find context by exact match or partial match
	var selectedContext string
	var matchedContexts []string

	for _, ctx := range cfg.Contexts {
		if ctx.Name == contextName {
			selectedContext = ctx.Name
			break
		}
		if strings.Contains(ctx.Name, contextName) {
			matchedContexts = append(matchedContexts, ctx.Name)
		}
	}

	// If no exact match, check for partial matches
	if selectedContext == "" {
		if len(matchedContexts) == 0 {
			fmt.Fprintf(os.Stderr, "%sError: Context '%s' not found%s\n", ui.Red, contextName, ui.Reset)
			fmt.Fprintf(os.Stderr, "\nAvailable contexts:\n")
			for _, ctx := range cfg.Contexts {
				marker := " "
				if ctx.Name == cfg.CurrentContext {
					marker = ui.BoldGreen + "●" + ui.Reset
				}
				fmt.Fprintf(os.Stderr, "  %s %s\n", marker, ctx.Name)
			}
			os.Exit(1)
		}
		if len(matchedContexts) == 1 {
			selectedContext = matchedContexts[0]
			fmt.Printf("%sUsing context: %s%s\n", ui.Gray, selectedContext, ui.Reset)
		} else {
			fmt.Fprintf(os.Stderr, "%sError: Multiple contexts match '%s':%s\n", ui.Red, contextName, ui.Reset)
			for _, ctx := range matchedContexts {
				fmt.Fprintf(os.Stderr, "  - %s\n", ctx)
			}
			os.Exit(1)
		}
	}

	// Switch context
	if err := switchContext(selectedContext); err != nil {
		fmt.Fprintf(os.Stderr, "%sError switching context: %v%s\n", ui.Red, err, ui.Reset)
		os.Exit(1)
	}

	fmt.Printf("\n%s✓ Successfully switched to context: %s%s%s\n", ui.BoldGreen, ui.BoldCyan, selectedContext, ui.Reset)

	// Set namespace if provided
	if namespace != "" {
		if err := config.UpdateContextNamespace(configPath, selectedContext, namespace); err != nil {
			fmt.Fprintf(os.Stderr, "%sError updating namespace: %v%s\n", ui.Red, err, ui.Reset)
			os.Exit(1)
		}
		fmt.Printf("%s✓ Namespace set to: %s%s%s\n", ui.BoldGreen, ui.BoldCyan, namespace, ui.Reset)
	}

}

func handleInteractiveMode(cfg *config.KubeConfig, configPath string) {
	ui.DisplayContexts(cfg)

	choice, err := ui.PromptSelection(len(cfg.Contexts))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\nInvalid input%s\n", ui.Red, ui.Reset)
		os.Exit(1)
	}

	if choice == 0 {
		fmt.Printf("%s\nOperation cancelled%s\n", ui.Yellow, ui.Reset)
		return
	}

	if choice < 1 || choice > len(cfg.Contexts) {
		fmt.Fprintf(os.Stderr, "%s\nInvalid choice. Please select a number between 1 and %d%s\n", ui.Red, len(cfg.Contexts), ui.Reset)
		os.Exit(1)
	}

	selectedContext := cfg.Contexts[choice-1].Name

	// Switch context
	if err := switchContext(selectedContext); err != nil {
		fmt.Fprintf(os.Stderr, "%sError switching context: %v%s\n", ui.Red, err, ui.Reset)
		os.Exit(1)
	}

	fmt.Printf("\n%s✓ Successfully switched to context: %s%s%s\n", ui.BoldGreen, ui.BoldCyan, selectedContext, ui.Reset)

	// Ask if user wants to set namespace
	if ui.PromptNamespaceChoice() {
		namespaces, err := config.GetContextNamespaces(selectedContext)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%sError fetching namespaces: %v%s\n", ui.Red, err, ui.Reset)
			os.Exit(1)
		}

		currentNamespace := cfg.Contexts[choice-1].Context.Namespace
		if currentNamespace == "" {
			currentNamespace = "default"
		}

		ui.DisplayNamespaces(namespaces, currentNamespace)

		nsChoice, err := ui.PromptNamespaceSelection(len(namespaces))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\nInvalid input%s\n", ui.Red, ui.Reset)
			os.Exit(1)
		}

		var selectedNamespace string
		if nsChoice == 0 {
			// Custom namespace
			selectedNamespace, err = ui.PromptCustomNamespace()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\nInvalid input%s\n", ui.Red, ui.Reset)
				os.Exit(1)
			}
		} else if nsChoice > 0 && nsChoice <= len(namespaces) {
			selectedNamespace = namespaces[nsChoice-1]
		} else {
			fmt.Fprintf(os.Stderr, "%s\nInvalid choice%s\n", ui.Red, ui.Reset)
			os.Exit(1)
		}

		// Update namespace in kubeconfig
		err = config.UpdateContextNamespace(configPath, selectedContext, selectedNamespace)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%sError updating namespace: %v%s\n", ui.Red, err, ui.Reset)
			os.Exit(1)
		}

		fmt.Printf("\n%s✓ Namespace set to: %s%s%s\n", ui.BoldGreen, ui.BoldCyan, selectedNamespace, ui.Reset)
	}

}

func switchContext(contextName string) error {
	cmd := exec.Command("kubectl", "config", "use-context", contextName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", output)
		return err
	}
	return nil
}
