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
| 📊 **Tableau de bord interactif** | Jauges graphiques dans le terminal montrant l'espace disque occupé par les volumes, conteneurs, images et caches de build. |
| ☑️ **Sélection sélective** | Naviguez avec les flèches, cochez avec la <kbd>Espace</kbd> pour choisir précisément quels éléments passer à la trappe. |
| 📉 **Tri par taille** | Les plus gros consommateurs d'espace remontent automatiquement en tête de liste. |
| ⏰ **Alertes d'ancienneté** | Barnacle repère et signale les ressources inutilisées depuis plus de X jours, pour ne rien laisser s'incruster trop longtemps. |

---

## 🚀 Quick Start

Aucune installation nécessaire. Montez simplement le socket Docker et lancez :

```bash
docker run --rm -it -v /var/run/docker.sock:/var/run/docker.sock horacioskrp/barnacle
```

Le tableau de bord s'ouvre directement dans votre terminal. Naviguez, sélectionnez, grattez. 🐚

---

## 🎮 Utilisation

| Touche | Action |
|---|---|
| `↑` / `↓` | Se déplacer dans la liste |
| `Espace` | Cocher / décocher un élément |
| `Entrée` | Confirmer et supprimer la sélection |
| `s` | Trier par taille décroissante |
| `q` | Quitter Barnacle |

---

## 🤝 Contribuer

Les contributions sont les bienvenues à bord ! Voici le flux à suivre :

1. **Fork** le dépôt
2. Créez votre branche de fonctionnalité à partir de `develop` :
   ```bash
   git checkout develop
   git checkout -b feature/ma-nouvelle-fonctionnalite
   ```
3. Commitez vos changements avec des messages clairs
4. Poussez votre branche vers votre fork
5. Ouvrez une **Pull Request** vers la branche `develop` du dépôt principal

Toute PR doit cibler `develop`, jamais `main` directement. 🧭

---

## 📜 Licence

Distribué sous licence **MIT**. Voir [`LICENSE`](./LICENSE) pour plus de détails.

---

<p align="center">
  Fait avec 🐚 et beaucoup de café par <a href="https://github.com/horacioskrp">@horacioskrp</a><br/>
  Si Barnacle vous a fait gagner de la place, laissez une ⭐ !
</p>
