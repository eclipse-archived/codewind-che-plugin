# Codewind Che Plug-in
Use the Eclipse Codewind sidecar plug-in for Eclipse Che to enable Theia to communicate with the Codewind server container.

[![License](https://img.shields.io/badge/License-EPL%202.0-red.svg?label=license&logo=eclipse)](https://www.eclipse.org/legal/epl-2.0/)
[![Build Status](https://ci.eclipse.org/codewind/buildStatus/icon?job=Codewind%2Fcodewind-che-plugin%2Fmaster)](https://ci.eclipse.org/codewind/job/Codewind/job/codewind-che-plugin/job/master/)
[![Chat](https://img.shields.io/static/v1.svg?label=chat&message=mattermost&color=145dbf)](https://mattermost.eclipse.org/eclipse/channels/eclipse-codewind)

## What is the Eclipse Codewind sidecar container?
The Codewind sidecar container includes the following responsibilities:
- The sidecar deploys the Codewind server container.
    - The sidecar renders the deployment and service templates and applies them with the `kubectl apply` command.
    - When the workspace is shut down or deleted, the sidecar tears down Codewind and any deployed applications.
- The sidecar sets up a reverse proxy for the Theia extension.
    - Nginx is used for the proxy because it can handle both HTTP requests and socket.io.
    - The Theia plug-in communicates with the reverse proxy, which then forwards requests to Codewind. This chain of communication avoids the addition of code in the Theia plug-in to discover and manage the connection to Codewind.
- The sidecar runs the `filewatcherd` daemon to track user code changes.
    - The `filewatcherd` daemon watches for changes in each user's project and communicates with Codewind, letting it know to start a build if required.
    - For more information on `filewatcherd`, see [eclipse/codewind-filewatchers](https://github.com/eclipse/codewind-filewatchers).

## Installing Codewind on Eclipse Che

To install Codewind on Eclipse Che, please consult [Installing and Using Codewind on Kubernetes](https://www.eclipse.org/codewind/installoncloud.html)

## Developing

### Prerequisites

- Install Docker 17.05 or later.

### Building

To build the sidecar image, run `./build.sh`.

### Deploying

For instructions on deploying custom builds of the Codewind Che plugin, consult DEVELOPING.md

## Contributing
We use the main Codewind git repo (https://github.com/eclipse/codewind) for issue tracking.

Submit issues and contributions:
1. [Submitting issues](https://github.com/eclipse/codewind/issues)
2. [Contributing](CONTRIBUTING.md)
