# Stellaris Patch

[Stellaris](https://store.steampowered.com/app/281990/Stellaris/) base game patcher for modding

> **Requires a legal installation of Stellaris.**

## Requirements

- [Stellaris](https://store.steampowered.com/app/281990/Stellaris/) base game
- [Git Bash](https://git-scm.com) for applying/generating patches on Windows

## CLI Usage

### Global flags
- `--base` Base game directory, defaulting to `STELLARIS_HOME` environment variable

### Commands

- `init-workspace` Create a new workspace
- `apply-patch` Apply patches in current workspace, add file names (relative path of base game directory) following to import new files from base game
- `rebuild-patch` Rebuild patches for changed files in current workspace, add file names (relative path of `src/` directory) following to force regenerate
- `deploy` Install to mods directory
  - `--target` Target directory, defaulting to `~/Documents/Paradox Interactive/Stellaris/mods` 
  - `--name` Target mod name, default `stellaris_patch`
  - `--purge` Purge existing mod only
- `install` Install to game directory
  - `--purge` Recover original game files, recommended before base game update if you have run `install` before
  - `--force` Force mode for purge, won't check if installed
