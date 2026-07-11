# 🏗️ Architecture

Barnacle suit l'organisation standard d'un projet Go : un binaire fin dans `cmd/`, et toute la logique dans des packages privés sous `internal/`.

```
barnacle/
├── cmd/
│   └── barnacle/
│       └── main.go       # Point d'entrée : connexion Docker + démarrage du programme Bubble Tea
├── internal/
│   ├── docker/
│   │   ├── docker.go      # Toute la logique métier Docker (analyse + nettoyage)
│   │   └── docker_test.go
│   └── tui/
│       ├── ui.go          # Interface terminal (TUI) au format Model-View-Update
│       └── ui_test.go
├── docs/                  # Documentation détaillée (ce dossier)
├── go.mod
└── go.sum
```

Le préfixe `internal/` empêche tout module externe d'importer ces packages : ce sont des détails d'implémentation propres à Barnacle, pas une bibliothèque publique.

## `internal/docker` — Couche Docker

- **`docker.Client`** encapsule `*client.Client` (SDK `github.com/moby/moby/client`) et se connecte via `client.FromEnv`, ce qui respecte les variables d'environnement standard (`DOCKER_HOST`, `DOCKER_CERT_PATH`, TLS...) et retombe par défaut sur `/var/run/docker.sock`.
- **`Analyze(ctx)`** appelle `DiskUsage` en mode verbeux et découpe le résultat en quatre `Category` :
  - `CategoryDanglingImages` — images sans tag (`<none>:<none>`)
  - `CategoryStoppedContainers` — conteneurs dans un état `exited`, `created` ou `dead`
  - `CategoryOrphanVolumes` — volumes dont `UsageData.RefCount == 0`
  - `CategoryBuildCache` — enregistrements de cache de build non utilisés (`InUse == false`)
- Chaque catégorie calcule aussi l'âge de son élément le plus ancien. Si cet âge dépasse `staleThreshold` (7 jours), `Category.HasStale` passe à `true` — c'est l'alerte intelligente affichée dans l'UI.
- **`Prune(ctx, selected)`** ne nettoie que les catégories cochées par l'utilisateur, une par une, et continue même si l'une d'elles échoue (les erreurs sont collectées dans `PruneSummary`, jamais silencieusement ignorées ni fatales pour les autres étapes).

## `internal/tui` — Couche TUI (Bubble Tea)

Architecture **Model-View-Update**, exposée via `tui.Model` et `tui.NewModel(*docker.Client)` :

| État (`sessionState`) | Déclencheur | Écran affiché |
|---|---|---|
| `stateLoading` | Démarrage du programme | Message d'analyse en cours |
| `stateBrowsing` | Réception de `diskUsageMsg` | Jauge + liste des 4 catégories avec cases à cocher |
| `stateConfirming` | Touche `entrée` avec ≥1 sélection | Récapitulatif de la sélection, avertissement renforcé si des volumes sont concernés |
| `statePruning` | Touche `y`/`entrée` sur l'écran de confirmation | Message de nettoyage en cours |
| `stateSummary` | Réception de `pruneResultMsg` | Récapitulatif de l'espace libéré par catégorie |
| `stateError` | Réception de `errMsg` | Message d'erreur, sortie possible |

Les appels réseau vers le démon Docker (`Analyze`, `Prune`) sont toujours exécutés dans des `tea.Cmd` (fonctions asynchrones), jamais directement dans `Update`, pour ne pas bloquer le rendu du terminal. Le package `tui` dépend de `internal/docker` (jamais l'inverse) : la couche Docker ignore tout de l'interface qui la pilote.

Le rendu (`View`) est entièrement délégué à Lipgloss : couleurs, bordures, mise en forme de la jauge `[████░░░░]` et des lignes de catégorie.

## `cmd/barnacle` — Point d'entrée

Séquence stricte :

1. `docker.NewClient()` — si le socket n'est pas accessible, le programme s'arrête immédiatement avec un message clair sur `stderr` (pas de crash silencieux).
2. `tea.NewProgram(tui.NewModel(dockerClient), tea.WithAltScreen())` — lance l'interface en plein écran alterné (restaure le terminal à la sortie).
3. Toute erreur d'exécution du programme Bubble Tea est remontée et fait sortir avec un code non nul.
