package ui

import (
	"fmt"

	"kc/internal/config"
)

// DisplayContexts shows all available contexts with the current one highlighted
func DisplayContexts(cfg *config.KubeConfig) {
	fmt.Println()
	fmt.Printf("%s╭─────────────────────────────────────────────────────────╮%s\n", BoldCyan, Reset)
	fmt.Printf("%s│               Kubernetes Context Switcher               │%s\n", BoldCyan, Reset)
	fmt.Printf("%s╰─────────────────────────────────────────────────────────╯%s\n", BoldCyan, Reset)
	fmt.Println()

	for i, ctx := range cfg.Contexts {
		marker := " "
		if ctx.Name == cfg.CurrentContext {
			marker = BoldGreen + "●" + Reset
			fmt.Printf(" %s [%s%d%s] %s%s%s\n",
				marker, Yellow, i+1, Reset, BoldGreen, ctx.Name, Reset)
		} else {
			fmt.Printf(" %s [%s%d%s] %s\n",
				marker, Yellow, i+1, Reset, ctx.Name)
		}
		fmt.Printf("%s      ├─ cluster: %s%s\n", Gray, ctx.Context.Cluster, Reset)
		fmt.Printf("%s      └─ user: %s%s\n", Gray, ctx.Context.User, Reset)
		if i < len(cfg.Contexts)-1 {
			fmt.Println()
		}
	}
}

// PromptSelection prompts the user to select a context
func PromptSelection(max int) (int, error) {
	fmt.Println()
	fmt.Printf("%s─────────────────────────────────────────────────────────%s\n", BoldCyan, Reset)
	fmt.Printf("%sSelect context [1-%d] (0 to cancel): %s", Blue, max, Reset)
	var choice int
	_, err := fmt.Scanln(&choice)
	return choice, err
}

// PromptNamespaceChoice asks if user wants to set a namespace
func PromptNamespaceChoice() bool {
	fmt.Println()
	fmt.Printf("%s─────────────────────────────────────────────────────────%s\n", BoldCyan, Reset)
	fmt.Printf("%sDo you want to set a namespace? [y/N]: %s", Blue, Reset)

	var response string
	fmt.Scanln(&response)

	return response == "y" || response == "Y"
}

// DisplayNamespaces shows available namespaces
func DisplayNamespaces(namespaces []string, currentNamespace string) {
	fmt.Println()
	fmt.Printf("%s╭─────────────────────────────────────────────────────────╮%s\n", BoldCyan, Reset)
	fmt.Printf("%s│                   Available Namespaces                  │%s\n", BoldCyan, Reset)
	fmt.Printf("%s╰─────────────────────────────────────────────────────────╯%s\n", BoldCyan, Reset)
	fmt.Println()

	for i, ns := range namespaces {
		marker := " "
		if ns == currentNamespace {
			marker = BoldGreen + "●" + Reset
			fmt.Printf(" %s [%s%d%s] %s%s%s\n", marker, Yellow, i+1, Reset, BoldGreen, ns, Reset)
		} else {
			fmt.Printf(" %s [%s%d%s] %s\n", marker, Yellow, i+1, Reset, ns)
		}
	}

	// Option to enter custom namespace
	fmt.Printf("\n %s [%s0%s] %sEnter custom namespace%s\n", " ", Yellow, Reset, Gray, Reset)
}

// PromptNamespaceSelection prompts for namespace selection
func PromptNamespaceSelection(max int) (int, error) {
	fmt.Println()
	fmt.Printf("%s─────────────────────────────────────────────────────────%s\n", BoldCyan, Reset)
	fmt.Printf("%sSelect namespace [1-%d] (0 for custom): %s", Blue, max, Reset)
	var choice int
	_, err := fmt.Scanln(&choice)
	return choice, err
}

// PromptCustomNamespace prompts for a custom namespace name
func PromptCustomNamespace() (string, error) {
	fmt.Printf("%sEnter namespace name: %s", Blue, Reset)
	var namespace string
	_, err := fmt.Scanln(&namespace)
	return namespace, err
}
