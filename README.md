# 🐚 Barnacle

**Grattez les bernacles qui alourdissent votre Docker.**

Barnacle est un outil **TUI (Terminal User Interface)** interactif qui vous aide à repérer et supprimer, en un clin d'œil, tout ce qui s'accroche inutilement à votre installation Docker : volumes orphelins, conteneurs arrêtés, images `<none>` et caches de build obsolètes.

Comme la coque d'un navire qui traîne des bernacles au fil du temps, votre Docker accumule des couches mortes qui le ralentissent et saturent votre disque. Barnacle vous donne le grattoir 🪸.

<p align="center">
  <img src="https://img.shields.io/badge/status-en%20d%C3%A9veloppement-orange?style=for-the-badge" alt="Status" />
  <img src="https://img.shields.io/github/license/horacioskrp/barnacle?style=for-the-badge&color=blue" alt="License" />
  <img src="https://img.shields.io/github/go-mod/go-version/horacioskrp/barnacle?style=for-the-badge" alt="Go Version" />
  <img src="https://img.shields.io/github/stars/horacioskrp/barnacle?style=for-the-badge&color=yellow" alt="GitHub Stars" />
</p>

---

## ⚓ Pourquoi Barnacle ?

`docker system prune` fait le ménage à l'aveugle. Vous ne savez jamais vraiment **ce que vous vous apprêtez à perdre**.

Barnacle inverse la logique : il vous montre l'état de votre coque Docker en temps réel, dans un joli tableau de bord de terminal, et vous laissez décider **exactement** ce qui part à la casse.

- 🧹 Nettoyage **ciblé**, pas à l'aveugle
- 📊 Visualisation claire de l'espace disque récupérable
- 🛡️ Aucune suppression sans confirmation explicite de votre part
- 🐳 Zéro installation — tourne dans un simple conteneur

---

## ✨ Fonctionnalités

| | |
|---|---|
| 📊 **Tableau de bord interactif** | Jauge graphique dans le terminal montrant l'espace récupérable, catégorie par catégorie (images suspendues, conteneurs arrêtés, volumes orphelins, cache de build). |
| ☑️ **Sélection sélective** | Naviguez avec les flèches (ou `j`/`k`), cochez avec la <kbd>Espace</kbd> pour choisir précisément quoi nettoyer. |
| 📐 **Taille réelle par catégorie** | Chaque ligne affiche la taille exacte et le nombre d'éléments concernés (ex : `Images suspendues 4.2 GB (7 éléments)`). |
| ⏰ **Alertes d'ancienneté** | Barnacle signale les ressources inutilisées depuis plus de 7 jours avec `⚠ inutilisé depuis X jours`. |

---

## 🚀 Quick Start

Aucune installation nécessaire. Montez simplement le socket Docker et lancez :

```bash
docker run --rm -it -v /var/run/docker.sock:/var/run/docker.sock horacioskrp/barnacle
```

Le tableau de bord s'ouvre directement dans votre terminal. Naviguez, sélectionnez, grattez. 🐚

---

## 📁 Structure du projet

```
barnacle/
├── main.go     # Point d'entrée : connexion Docker + démarrage du programme
├── docker.go   # Logique métier Docker (analyse d'espace + nettoyage ciblé)
├── ui.go       # Interface terminal (Bubble Tea / Lipgloss)
├── docs/       # Documentation détaillée
├── go.mod
└── go.sum
```

---

## 📚 Documentation

- [Guide d'utilisation](./docs/USAGE.md) — écrans, raccourcis clavier, prérequis.
- [Architecture](./docs/ARCHITECTURE.md) — organisation du code et logique interne.
- [Guide de contribution](./docs/CONTRIBUTING.md) — flux Git et checklist avant PR.

---

## 🎮 Raccourcis clavier

| Touche | Action |
|---|---|
| `↑` / `k` | Se déplacer vers le haut |
| `↓` / `j` | Se déplacer vers le bas |
| `Espace` | Cocher / décocher un élément |
| `Entrée` | Passer à l'écran de confirmation |
| `y` / `n` | Confirmer / annuler le nettoyage |
| `q` / `Ctrl+C` | Quitter Barnacle |

Détails complets dans [docs/USAGE.md](./docs/USAGE.md).

---

## 🤝 Contribuer

Fork → branche depuis `develop` → Pull Request vers `develop`. Le détail complet du flux est dans [docs/CONTRIBUTING.md](./docs/CONTRIBUTING.md).

---

## 📜 Licence

Distribué sous licence **MIT**. Voir [`LICENSE`](./LICENSE) pour plus de détails.

---

<p align="center">
  Fait avec 🐚 et beaucoup de café par <a href="https://github.com/horacioskrp">@horacioskrp</a><br/>
  Si Barnacle vous a fait gagner de la place, laissez une ⭐ !
</p>
