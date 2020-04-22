#!/bin/bash
#
#*******************************************************************************
# Copyright (c) 2019 IBM Corporation and others.
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Eclipse Public License v2.0
# which accompanies this distribution, and is available at
# http://www.eclipse.org/legal/epl-v20.html
#
# Contributors:
#     IBM Corporation - initial API and implementation
#*******************************************************************************

# Pushes the Nginx sidecar image up to the specified registry

set -e

if [ "$#" -lt 1 ]; then
    echo "usage: $0 <registry-to-push-sidecar-to> <image-tag>"
    exit 1
fi

REGISTRY=$1

# Set the image tag
if [ -z "$2" ]; then
    TAG=latest
else
    TAG=$2
fi

SCRIPTS_DIR="$(dirname $0)"
BASE_DIR="$(dirname $SCRIPTS_DIR)"

docker tag codewind-che-sidecar $REGISTRY/codewind-che-sidecar:$TAG
docker push $REGISTRY/codewind-che-sidecar:$TAG

# Create the meta.yaml for the Sidecar container
mkdir -p $BASE_DIR/publish/codewind-sidecar/$TAG
cat <<EOF > publish/codewind-sidecar/$TAG/meta.yaml
id: codewind-sidecar
apiVersion: v2
version: $TAG
type: Che Plugin
name: CodewindPlugin
title: CodewindPlugin
description: Enables iterative development and deployment in Che
icon: https://raw.githubusercontent.com/eclipse/codewind-vscode/master/dev/res/img/codewind.png
publisher: eclipse
repository: https://github.com/eclipse/codewind-che-plugin
category: Other
firstPublicationDate: "2019-05-30"
latestUpdateDate: "$(date '+%Y-%m-%d')"
spec:
  containers:
  - name: codewind-che-sidecar
    image: $REGISTRY/codewind-che-sidecar:$TAG
    volumes:
      - mountPath: "/projects"
        name: projects
    ports:
      - exposedPort: 9090
EOF

# Create the meta.yaml for the Theia extension
mkdir -p $BASE_DIR/publish/codewind-theia/$TAG
cat <<EOF > publish/codewind-theia/$TAG/meta.yaml
apiVersion: v2
publisher: eclipse
name: codewind-plugin
version: $TAG
type: VS Code extension
displayName: Codewind VS Code Extension
title: Codewind Extension for VS Code
description: Codewind Extension for Theia
icon: https://raw.githubusercontent.com/eclipse/codewind-vscode/master/dev/res/img/codewind.png
repository: http://github.com/eclipse/codewind-vscode/
category: Other
firstPublicationDate: "2019-05-30"
latestUpdateDate: "$(date '+%Y-%m-%d')"
spec:
  extensions:
    - http://archive.eclipse.org/codewind/codewind-vscode/master/latest/codewind-theia.vsix
EOF

echo "Published the codewind-sidecar and codewind-theia meta.yamls under $BASE_DIR/publish"
