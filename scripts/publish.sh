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
set -u

echo "$PASSWORD" | docker login -u "$USERNAME" --password-stdin $REGISTRY
docker tag codewind-che-sidecar $REGISTRY/codewind-che-sidecar
docker push $REGISTRY/codewind-che-sidecar