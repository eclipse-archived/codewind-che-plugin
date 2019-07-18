## Building & Using the Codewind Plugin on Che

### Build the Codewind plugin

Run the script build.sh to build the Codewind Che sidecar.

### Tag and Push the Codewind plugin

1. Tag the sidecar:
   ```
   docker tag codewind-che-sidecar ${registry}/codewind-che-sidecar
   ```

2. Push the sidecar up to a registry:
   ```
   docker push ${registry}/codewind-che-sidecar
   ```

### Deploying a Custom Codewind Che sidecar

1. If modifying the sidecar image, create a `meta.yaml` for the Codewind sidecar plugin, and host it publicly. Make sure to also set the container image to the image that you pushed up in the earlier step. For example:

```
id: codewind-sidecar
apiVersion: v2
version: latest
type: Che Plugin
name: CodewindPlugin
title: CodewindPlugin
description: Enables iterative development and deployment in Che
icon: https://raw.githubusercontent.com/eclipse/codewind-vscode/master/dev/res/img/codewind.png
publisher: Eclipse
repository: https://github.com/eclipse/codewind-che-plugin
category: Other
firstPublicationDate: "2019-05-30"
latestUpdateDate: "2019-06-26"
spec:
  containers:
  - name: codewind-che-sidecar
    image: ${REGISTRY}/codewind-che-sidecar:latest
    volumes:
      - mountPath: "/projects"
        name: projects
    ports:
      - exposedPort: 9090
```

2. If modifying the Codewind theia extension, create a `meta.yaml` for the Codewind theia extension, and host it publicly. Make sure to also link directly to your Theia extension:
```
apiVersion: v2
publisher: Eclipse
name: codewind-plugin
version: latest
type: VS Code extension
displayName: Codewind VS Code Extension
title: Codewind Extension for VS Code
description: Codewind Extension for Theia
icon: https://raw.githubusercontent.com/eclipse/codewind-vscode/master/dev/res/img/codewind.png
repository: http://github.com/eclipse/codewind-vscode/
category: Other
firstPublicationDate: "2019-05-30"
latestUpdateDate: "2019-06-26"
spec:
  extensions:
    - ${SERVER}/codewind-theia-0.2.0.vsix
```

3. Finally, to create a Che Codewind workspace, write a dev file and host it publicly, making sure to set the links to the codewind-sidecar and codewind-theia meta.yamls as needed (link to your custom meta.yamls).

```
apiVersion: 1.0.0
metadata:
  name: codewind-che
projects:
  - name: goproj
    source:
      type: git
      location: 'https://github.com/microclimate-dev2ops/microclimateGoTemplate'
components:
  - alias: theia-ide
    type: cheEditor
    id: eclipse/che-theia/7.0.0-rc-3.0
  - alias: codewind-sidecar
    type: chePlugin
    id: https://raw.githubusercontent.com/eclipse/codewind-che-plugin/master/plugins/codewind/codewind-sidecar/latest/meta.yaml
  - alias: codewind-theia
    type: chePlugin
    id: https://raw.githubusercontent.com/eclipse/codewind-che-plugin/master/plugins/codewind/codewind-theia/latest/meta.yaml
```
  
  Then create the workspace in Che by accessing http://$CHE_DOMAIN/f?url=${DEVFILE_LINK} in your browser, where ${DEVFILE_LINK} is the direct link to the devfile you created.