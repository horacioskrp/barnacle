package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	docker, err := NewDockerClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "barnacle: %v\n", err)
		os.Exit(1)
	}
	defer docker.Close()

	program := tea.NewProgram(initialModel(docker), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "barnacle: erreur d'exécution du programme: %v\n", err)
		os.Exit(1)
	}
}
