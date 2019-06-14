## Building & Using the Codewind Plugin on Che

### Build the Codewind plugin

Run the script build.sh to build the Codewind Che sidecar.

### Creating a Che Codewind workspace

To create a Che Codewind workspace, write a dev file and host it publicly with the following references to the Che sidecar and the Theia plugin metadata

```
---
specVersion: 0.0.1
name: codewind-plugin-env
projects:
  - name: goproj
    source:
      type: git
      location: 'https://github.com/microclimate-dev2ops/microclimateGoTemplate'
components:
  - alias: theia-ide
    type: cheEditor
    id: eclipse/che-theia/next
  - alias: exec-plugin
    type: chePlugin
    id: eclipse/che-machine-exec-plugin/0.0.1
  - alias: codewind-plugin
    type: chePlugin
    id: {server}/plugins/codewind/codewind-plugin/0.0.1/meta.yaml
  - alias: codewind-theia
    type: chePlugin
    id: {server}/plugins/codewind/codewind-theia/0.0.1/meta.yaml
```

The `codewind-plugin` meta.yaml will provide the details of the Che sidecar. You must have built and pushed the sidecar image up to a docker registry

```
id: codewind-plugin
apiVersion: v2
version: 0.0.1
type: Che Plugin
name: codewind-plugin
title: codewind-plugin
description: Enables iterative development and deployment in Che
icon: https://raw.githubusercontent.com/IBM/charts/master/logo/microclimate-logo.png
publisher: IBM
category: Other
repository: {server}/plugins/codewind/codewind-plugin
firstPublicationDate: "2019-02-20"
spec:
  containers:
  - name: codewind-che-sidecar
    image: {REGISTRY_URL}/codewind-che-sidecar
    volumes:
      - mountPath: "/projects"
        name: projects
    ports:
      - exposedPort: 9090
```

The `codewind-theia` meta.yaml will provide the details to the Theia plugin, where the `codewind_plugin.theia` is hosted

```
apiVersion: v2
publisher: IBM
name: codewind-plugin
version: 0.0.1
type: Theia plugin
displayName: Codewind Theia Plugin
title: Codewind Plugin for Theia
description: Codewind Plugin for Theia
icon: https://raw.githubusercontent.com/vitaliy-guliy/che-theia-plugin-registry/master/icons/tree.png
repository: https://github.com/eclipse/che-theia-samples/tree/master/samples/hello-world-frontend-plugin
category: Other
firstPublicationDate: "2019-03-13"
latestUpdateDate: "2019-04-09"
spec:
  extensions:
    - {server}/plugins/codewind/codewind-theia/0.0.1/codewind_plugin.theia
```
