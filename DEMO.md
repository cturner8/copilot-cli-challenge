# binmate demo/screenshots

## Demo

Short demo recording:

![](./images/demo.mov)

This demonstrates the following functionality:
- first check current version of the `gh` CLI (initially not installed)
- launch `binmate`
- install non-latest version
- verify installed version
- re-launch `binmate`
- update to latest version
- verify latest version is installed

## Screenshots



### Binary List View

![](./images/binary-list-view.png)

Provides an overview of configured binaries with actions for:
- view installed versions
- search
- filter
- sort
- add new binary
- install new version
- update selected binary
- remove selected binary

### Binary List View with Search

![](./images/binary-list-view-with-search.png)

Enter search query performs an in memory search of the displayed binaries.

### Binary List View Filter Panel

![](./images/binary-list-view-filter-panel.png)

Filter panel provides the following options:
- provider (currently only github is supported)
- format
- status (installed/not installed)

### Binary List View with Filter

![](./images/binary-list-view-with-filter.png)

### Binary Add View

![](./images/binary-add-view.png)

Provides input for entering a github release URL. Entered URL is parsed and required metadata extracted, redirects to [Binary Install Configuration View](#binary-install-configuration-view) to allow for refinement of extracted metadata.

### Binary Add Configuration View

![](./images/binary-add-config-view.png)

Displays metadata extracted from entered URL in [Binary Install View](#binary-install-view) with option to override any of the identified values.

![](./images/binary-add-success-view.png)

Success prompt following save.

### Binary Installed Versions View

![](./images/binary-installed-versions-view.png)

Provides an overview of the selected binary details along with a summary of the installed versions.

Actions for the following:
- switch active version
- install new version
- check for version updates
- update version
- delete version
- view version release notes
- view associated repository info
- view available remote versions

### Binary Install View

![](./images/binary-install-view.png)

Provides input for version to be installed. Defaults to latest if not provided.

Can also be reached via the available versions view where the version input is automatically pre-populated.

### Binary Version Installed

![](./images/binary-version-installed.png)

Confirmation of successful version install.

### Binary Version Update Available

![](./images/binary-version-update-available.png)

Display of update available message following check for update action.

### Binary Version Update Installed

![](./images/binary-version-updated.png)

Confirmation prompt following a successful update operation.

### Binary Installed Version Switch

![](./images/binary-version-switch.png)

### Binary Available Versions View

![](./images/binary-available-versions-view.png)

Overview of available versions for chosen binary based on github releases.

Supports both pre-release and standard releases.

Provides the following actions:
- view release notes for selected version
- install selected version (jumps to binary install view with pre-populated version input)

### Binary Release Notes View

![](./images/binary-version-release-notes-view.png)

Overview of selected release version.

Available via the following locations:
- installed versions view
- available versions view

### Binary Repository Info View

![](./images/binary-repo-info.png)

Provides an overview of the github repository associated with a binary with star action.

> **Note**: as starring is an authenticated operation, a GitHub token must be available in your terminal environment.

### Binary Repository Star Action

![](./images/binary-repo-star-action.png)

Feedback provided for successful star operation.

> **Note**: depending on your type of access token, may not work on public repositories where you are not the owner.

