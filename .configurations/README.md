# Configuration for Development on Windows Using VS Code and Dev Containers

Radius can be developed on [multiple platforms using multiple configurations](docs/contributing/contributing-code/contributing-code-prerequisites). This folder contains automation for installing one such configuration: Visual Studio Code and dev containers running on Windows. The automation uses winget to install a baseline set of tools, including:

- Visual Studio Code
- Windows Subsystem for Linux (WSL), with the default Ubuntu instance
- Docker Desktop

Note: please review the Docker Desktop license terms before installing.

You can use these tools to run the Radius dev container on Windows.

## Prerequisites

- `winget` version 1.6 or greater
- PowerShell

## How to run the winget configuration on Windows

1. Open a PowerShell terminal window. The terminal instance does not have to run as administrator, but you will have to respond to interactive UAC prompts to perform some administrative actions.
1. Run the `Set-WingetConfiguration.ps1` script.

## Additional setup

After the winget configuration finishes, do these manual configuration steps.

- Reboot Windows to ensure Docker Desktop completes its setup.
- Ensure that you have the [Remote Development](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.vscode-remote-extensionpack) extension pack installed in VS Code.
- Launch Ubuntu and run the one-time setup steps to create the default user.
- Configure git on Ubuntu.
- Configure Docker Desktop to integrate with Ubuntu.
- Clone the Radius repo to a folder in WSL, and open the repo folder in Visual Studio Code. You can `cd` to the repo folder, then run this command to launch VS Code: `code .` (Be sure to include the `.` character in the command to tell VS Code to open that folder.)

After you open the Radius git repo in VS Code, you may be prompted to rebuild the development container. Do this step. If you are not prompted to rebuild the container, do the following:

1. Ensure that you have opened the Radius repo at the root folder of the repo (not the parent folder or a child folder).
1. Manually rebuld and run the dev container by running this command in VS Code: `Dev Containers: Rebuild and Reopen in Container`
