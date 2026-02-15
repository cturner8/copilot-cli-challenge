*This is a submission for the [GitHub Copilot CLI Challenge](https://dev.to/challenges/github-2026-01-21)*

## What I Built
<!-- Provide an overview of your application and what it means to you. -->

**[binmate](https://github.com/cturner8/copilot-cli-challenge)** is a CLI/TUI application for managing binary installations from GitHub releases. It provides an easy way to install, manage, and switch between different versions of command-line tools.

### Key Features

- **Interactive TUI**: Browse and manage binaries with a Terminal User Interface
- **CLI Commands**: Automate binary management with command-line interface
- **Version Management**: Install multiple versions and switch between them
- **GitHub Integration**: Automatically fetch releases from GitHub repositories
- **Database Tracking**: SQLite database tracks all installations and versions
- **Checksum Verification**: Ensures integrity of downloaded binaries

### Tech Stack

- Go
- Bubble Tea
- Lip Gloss
- Cobra
- Viper
- SQLite

[GitHub Repository](https://github.com/cturner8/copilot-cli-challenge)

## Demo
<!-- Share a link to your project and include a video walkthrough or screenshots showing your application in action. -->

## My Experience with GitHub Copilot CLI
<!-- Explain how you used GitHub Copilot CLI while building your project and how it impacted your development experience. -->

Development of the application used the following GitHub Copilot CLI capabilities:

- **Agent instructions** for baseline codebase structure and standards, later extended to also heavily utilise [beads](https://github.com/steveyegge/beads) for improved task tracking and persistent agent memory.
- **Custom agents** to provide more focused and specialised agents. The following were originally sourced from the [awesome-copilot](https://github.com/github/awesome-copilot/blob/main/docs/README.agents.md) repository:
    - Critical Thinking, see [critical-thinking.agent.md](https://github.com/cturner8/copilot-cli-challenge/blob/main/.github/agents/critical-thinking.agent.md)
    - Go Development Expert, revised from "Go MCP Server Development Expert" to be a generic go agent, see [go-expert.agent.md](https://github.com/cturner8/copilot-cli-challenge/blob/main/.github/agents/go-expert.agent.md)
    - SQLite Database Administrator, revised from "PostgreSQL Database Administrator" agent, see [sqlite-dba.agent.md](https://github.com/cturner8/copilot-cli-challenge/blob/main/.github/agents/sqlite-dba.agent.md)
- **MCP**
    - `context7` for providing up to date documentation for packages and libraries.
    - The `GitHub MCP` built into Copilot CLI was ideal for allowing the agent direct access to search public codebases, access to action workflow logs for troubleshooting, allowing follow up tasks and recommendations to be extracted from PR comments.
- **Plan mode** to refine and improve implementation plans for new features before work begins.
- **Delegate to cloud agent** once an implementation plan is finalised, allowing Copilot to continue working on a task freeing me up to focus on something else. It was also useful in allowing development to continue in moments where I had less time available to spend at my desk.
- **Code review** to identify any potential issues or areas for improvement in code changes, particularly helpful given I worked alone on the project.

<!-- Don't forget to add a cover image (if you want). -->
