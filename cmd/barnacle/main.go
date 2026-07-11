// Command barnacle launches the Barnacle terminal UI for analyzing and
// cleaning unused Docker resources.
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"barnacle/internal/docker"
	"barnacle/internal/tui"
)

func main() {
	dockerClient, err := docker.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "barnacle: %v\n", err)
		os.Exit(1)
	}
	defer dockerClient.Close()

	program := tea.NewProgram(tui.NewModel(dockerClient), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "barnacle: erreur d'exécution du programme: %v\n", err)
		os.Exit(1)
	}
}
