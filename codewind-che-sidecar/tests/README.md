# Codewind Sidecar Tests
Tests for verifying Codewind and sidecar container functionality in a Che environment.

[![License](https://img.shields.io/badge/License-EPL%202.0-red.svg?label=license&logo=eclipse)](https://www.eclipse.org/legal/epl-2.0/)
[![Jenkins](https://img.shields.io/static/v1.svg?label=builds&message=Jenkins&color=d24939&logo=jenkins&logoColor=ffffff)](https://ci.eclipse.org/codewind/job/Codewind/job/codewind-che-plugin/)
[![Chat](https://img.shields.io/static/v1.svg?label=chat&message=mattermost&color=145dbf)](https://mattermost.eclipse.org/eclipse/channels/eclipse-codewind)

## Prerequisites for installing and running the tests

- Bash shell environment
- Kubernetes cluster with Eclipse Che installed and pre-requisites for Codewind already set up (such as the cluster roles)
- User authenticated properly for the cluster
- `kubectl` tool configured for the cluster
- [BATS Bash testing environment](https://github.com/bats-core/bats-core)
- `curl` [tool](https://curl.haxx.se/)
- `jq` json parsing [tool](https://stedolan.github.io/jq/)
- `yq` yaml parsing [tool](https://github.com/mikefarah/yq)
 
## Running the tests

- To run the tests, invoke: `test_sidecar.sh <Che ingress domain> <Che namespace>`, for example `test_sidecar.sh che-che.10.98.130.246.nip.io che`
- The test suite will create a new Che workspace based on the latest Codewind devfile, test various scenarios, and finally delete the workspace
