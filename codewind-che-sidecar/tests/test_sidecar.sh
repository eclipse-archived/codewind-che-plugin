#!/usr/bin/env bash

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

if [[ ($# -ne 2) || ( -z "$1" ) || ( -z "$2" ) ]]; then
    echo -e "Usage: test_sidecar.sh <cluster IP> <Che namespace>"
    echo -e "Example: test_sidecar.sh 10.98.130.246 che"
    exit 1
fi

export CLUSTER_IP=$1
export CHE_NAMESPACE=$2

bats sidecarfvt.bats
