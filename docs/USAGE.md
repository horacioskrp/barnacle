# 📖 Guide d'utilisation

## Lancer Barnacle

Aucune installation nécessaire : montez le socket Docker et lancez le conteneur.

```bash
docker run --rm -it -v /var/run/docker.sock:/var/run/docker.sock horacioskrp/barnacle
```

Ou, en local depuis les sources :

```bash
go build -o barnacle .
./barnacle
```

## Écrans

1. **Analyse** — Barnacle interroge le démon Docker (`/system/df`) et calcule l'espace récupérable par catégorie.
2. **Tableau de bord** — une jauge affiche l'espace sélectionné par rapport au total récupérable, suivie de la liste des 4 catégories :
   - 🖼️ Images suspendues (`<none>:<none>`)
   - 📦 Conteneurs arrêtés
   - 💾 Volumes orphelins
   - 🧱 Cache de build obsolète
3. **Confirmation** — récapitulatif de ce qui va être supprimé (catégories, tailles, nombre d'éléments) et de l'espace total qui sera libéré. Si des volumes sont sélectionnés, un avertissement renforcé rappelle que l'action est irréversible et peut effacer des données.
4. **Nettoyage** — une fois confirmé, chaque catégorie sélectionnée est nettoyée l'une après l'autre.
5. **Résumé** — l'espace libéré est affiché par catégorie, avec le total.

## Raccourcis clavier

### Écran de sélection

| Touche | Action |
|---|---|
| `↑` / `k` | Monter dans la liste |
| `↓` / `j` | Descendre dans la liste |
| `Espace` | Cocher / décocher la catégorie sous le curseur |
| `Entrée` | Passer à l'écran de confirmation |
| `q` / `Ctrl+C` | Quitter |

### Écran de confirmation

| Touche | Action |
|---|---|
| `y` / `Entrée` | Confirmer et lancer le nettoyage |
| `n` / `Échap` | Annuler et revenir à la sélection |
| `q` / `Ctrl+C` | Quitter |

Pendant le nettoyage lui-même (`statePruning`), toutes les touches de sortie sont désactivées pour éviter d'interrompre une suppression en cours.

## Alertes de fraîcheur

Une catégorie affiche `⚠ inutilisé depuis X jours` lorsque son élément le plus ancien dépasse le seuil de 7 jours. Cela permet de repérer en un coup d'œil les ressources oubliées depuis longtemps, sans avoir à consulter les dates une par une.

## Prérequis

- Un démon Docker accessible (socket unix monté, ou `DOCKER_HOST` configuré).
- Les droits nécessaires pour lire et supprimer des ressources Docker (le conteneur doit tourner avec un utilisateur ayant accès au socket).
