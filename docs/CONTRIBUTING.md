# 🤝 Contribuer à Barnacle

Les contributions sont les bienvenues à bord ! Voici le flux à suivre :

1. **Fork** le dépôt.
2. Créez votre branche de fonctionnalité à partir de `develop` :
   ```bash
   git checkout develop
   git checkout -b feature/ma-nouvelle-fonctionnalite
   ```
3. Commitez vos changements avec des messages clairs et atomiques.
4. Vérifiez que le projet compile et passe les vérifications de base avant de pousser :
   ```bash
   go build ./...
   go vet ./...
   go test ./...
   gofmt -l .
   ```
5. Poussez votre branche vers votre fork.
6. Ouvrez une **Pull Request** vers la branche `develop` du dépôt principal — jamais directement vers `main`.

## Où intervenir

- [`internal/docker/docker.go`](../internal/docker/docker.go) : logique d'analyse et de nettoyage Docker.
- [`internal/tui/ui.go`](../internal/tui/ui.go) : interface terminal (Bubble Tea / Lipgloss).
- [`cmd/barnacle/main.go`](../cmd/barnacle/main.go) : point d'entrée du programme.

Voir [ARCHITECTURE.md](./ARCHITECTURE.md) pour comprendre l'organisation du code avant de contribuer.
