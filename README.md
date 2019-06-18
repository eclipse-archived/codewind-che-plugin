# Codewind Che Plug-in
Use the Eclipse Codewind sidecar plug-in for Eclipse Che to enable Theia to communicate with the Codewind server container.

[![License](https://img.shields.io/badge/License-EPL%202.0-red.svg?label=license&logo=eclipse)](https://www.eclipse.org/legal/epl-2.0/)

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


## Developing

### Prerequisites

- Install Docker 17.05 or later.

### Building

To build the sidecar image, run `./build.sh`.

### Deploying

For deployment instructions, see the README.md file at [eclipse/codewind-che-plugin](https://github.com/eclipse/codewind-che-plugin/tree/master/scripts).

## Contributing
Submit issues and contributions:
1. [Submitting issues](https://github.com/eclipse/codewind-che-plugin/issues)
2. [Contributing](CONTRIBUTING.md)