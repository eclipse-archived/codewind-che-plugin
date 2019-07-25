## Building & Using the Codewind Plugin on Che

### Build the Codewind plugin

Run the script build.sh to build the Codewind Che sidecar.

### Deploying a Custom Codewind Che sidecar

1. First, run the publish script located in this repository: `scripts/publish.sh $REGISTRY`, where `$REGISTRY` is a docker registry that you can push to, such as `docker.io/testuser` or `quay.io/testuser`.
    - The script will push the sidecar image up to the registry, and then generate the meta.yamls for `codewind-sidecar` and `codewind-theia` under the `publish/` folder.

2. Upload the meta.yamls somewhere publicly accessible, such as a GitHub repository.

3. Finally, to create a Che Codewind workspace, write a dev file and host it publicly, such as on Github. Make sure to set the URLs for `codewind-sidecar` and `codewind-theia` accordingly. If hosting the devfiles or plugins on GitHub, make sure you use the raw Github link (such as https://raw.githubusercontent.com/eclipse/codewind-che-plugin/master/devfiles/latest/devfile.yaml)

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
  
  Then create the workspace in Che by accessing http://$CHE_DOMAIN/f?url=${DEVFILE_LINK} in your browser, where ${DEVFILE_LINK} is the direct link to the devfile you created. Che will then create a workspace from that devfile.
  - An example of such a link is: http://che-eclipse-che.1.2.3.4.nip.io/f?url=https://raw.githubusercontent.com/eclipse/codewind-che-plugin/master/devfiles/latest/devfile.yaml


      