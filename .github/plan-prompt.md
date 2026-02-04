# binmate app

## Overview

binmate will be a CLI/TUI application for managing binary installations from remote repositories such as GitHub.

## Core Functionality

- Terminal User Interface (TUI) for main interactive interface
- Additional command line interface for non-interactive management operations (ideal for scripting/automation)
- Defaults to user level installations ($HOME directory in Linux environments)
- CLI command to import existing binaries installed manually
- Assets will be downloaded via REST API directly
- Initial focus will be on GitHub releases but should consider flexibility to extend to include additional providers in the future.
- Checksum verification to ensure integrity of installed assets
- Local SQLite database to track installed apps, versions and providers
- Installations will take place in a local cache/staging folder in home directory (potentially $HOME/.config/.binmate) with symlink to destination directory to allow for easily switching between installed versions and rollbacks (similar to something like nvm for Node.js installations)
- Installs should allow for resuming if a previous installation failed or was cancelled

## Tech Stack

- Go
- Cobra
- Viper
- SQLite
- Bubble Tea
- Lip Gloss
