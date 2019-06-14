# Codewind Che Plugin
The Codewind plugin for Eclipse Che!

[![Eclipse License](https://img.shields.io/badge/license-Eclipse-brightgreen.svg)](https://github.ibm.com/dev-ex/tempest/blob/master/LICENSE)

This repository contains the source code for the Eclipse Codewind sidecar plugin for Eclipse Che, allowing Theia to communicate with the Codewind server container.

The Codewind sidecar container has a number of responsibilities:
- Deploying the Codewind server container
    - It renders the deployment and service templates and `kubectl apply`'s them.
    - When the workspace is shut down or deleted, the sidecar will tear down Codewind, as well as any deployed applications.
- Setting up a reverse proxy for the Theia extension
    - We use Nginx for the proxy, as it can handle both HTTP requests and socket.io
    - The Theia plugin communicates with the reverse proxy (which then forwards requests to Codewind). This prevents us from having to add code in the Theia plugin to discover and manage the connection to Codewind
- Running the `filewatcherd` daemon, to track user code changes
    - `filewatcherd` watches for changes in each of the user's projects, and communicates with Codewind, letting it know to start a build (if required)
    - For more details on filewatcherd, see https://github.com/eclipse/codewind-filewatchers


## Development

### Prerequisites

- Docker 17.05 or higher

### Build

To build the sidecar image, run `./build.sh`.

### Deployment

Follow the readme at (https://github.com/eclipse/codewind-che-plugin/tree/master/scripts) for deployment instructions.

## Contributing
We welcome submitting issues and contributions.
1. [Submitting bugs](https://github.com/eclipse/codewind-che-plugin/issues)
2. [Contributing](CONTRIBUTING.md)

## License
[EPL 2.0](https://www.eclipse.org/legal/epl-2.0/)
