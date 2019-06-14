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

# Builds the nginx sidecar image and packages the Che plugin archive

set -e
set -u

BLUE='\033[1;34m'
RED='\031[1;34m'
NC='\033[0m'
DIR=$(cd "$(dirname "$0")"; pwd)

SCRIPTS_DIR=$(cd "$(dirname "$0")"; pwd)
BASE_DIR=$(dirname $SCRIPTS_DIR)

# Build the sidecar image
printf "${BLUE}Building the Codewind sidecar image${NC}\n"
cd ${BASE_DIR}/codewind-che-sidecar && ./build.sh
