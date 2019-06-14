#!/bin/bash

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

# Move to the directory of the script that included this include file
export SCRIPT_LOCT=$( cd $( dirname $0 ); pwd )
cd $SCRIPT_LOCT

export FROM_DOMAIN_NAME=localhost
export TO_DOMAIN_NAME=http://host.docker.internal:9090
export LISTEN_PORT=8080

docker stop codewind-che-sidecar >/dev/null 2>&1

docker rm -f codewind-che-sidecar >/dev/null 2>&1

# note when the side car container is launched within the che workspace, these parameters will have to be discovered in-container
docker run  --name codewind-che-sidecar -d \
	-p $LISTEN_PORT:$LISTEN_PORT \
	-e "_____FROM_DOMAIN_NAME=$FROM_DOMAIN_NAME" \
	-e "_____TO_DOMAIN_NAME=$TO_DOMAIN_NAME" \
	-e "_____LISTEN_PORT=$LISTEN_PORT" \
	--restart always \
	codewind-che-sidecar 

