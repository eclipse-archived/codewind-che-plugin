#!/usr/bin/env bats

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

# Load common utility functions
load testutil

# Setup code to run before each test
setup() {
    export time_before=$SECONDS

    if [ -z "$CHE_INGRESS_DOMAIN" ]; then
        echo "# Che ingress domain is not defined in " '$CHE_INGRESS_DOMAIN' >&3
        exit 1
    fi

    if [ -z "$CHE_NAMESPACE" ]; then
        echo "# Che namespace is not defined in " '$CHE_NAMESPACE' >&3
        exit 1
    fi

    if [ -z "$CLUSTER_IP" ]; then
        echo "# Cluster IP address is not defined in " '$CLUSTER_IP' >&3
        exit 1
    fi

    export CODEWIND_DEVFILE_URL=https://raw.githubusercontent.com/eclipse/codewind-che-plugin/master/devfiles/latest/devfile.yaml
    export KUBE_NAMESPACE_ARG="-n $CHE_NAMESPACE"

    # Discover workspace ID written into temporary file during workspace creation
    if [ -f che_workspace_id.txt ]; then
        export CHE_WORKSPACE_ID=$(cat che_workspace_id.txt)
    else
        export CHE_WORKSPACE_ID=workspace00000
    fi

    # Discover workspace pod and sidecar full names based on workspace ID
    export CHE_WORKSPACE_POD_FULLNAME=$(kubectl get pods -l che.original_name=che-workspace-pod --no-headers -o custom-columns=":metadata.name" $KUBE_NAMESPACE_ARG | grep $CHE_WORKSPACE_ID)
    export SIDECAR_CONTAINER_FULLNAME=$(kubectl get pods $CHE_WORKSPACE_POD_FULLNAME -o jsonpath='{.spec.containers[*].name}' $KUBE_NAMESPACE_ARG | sed 's/ /\n/g' | grep ^codewind-che-sidecar)

    # Set up Che access token for multi-user Che environment
    CHE_USER="admin"
    CHE_PASS="admin"
    KEYCLOAK_HOSTNAME=keycloak-"$CHE_NAMESPACE"."$CLUSTER_IP".nip.io
    TOKEN_ENDPOINT="http://${KEYCLOAK_HOSTNAME}/auth/realms/che/protocol/openid-connect/token" 
    export CHE_ACCESS_TOKEN=$(curl -sSL --data "grant_type=password&client_id=che-public&username=${CHE_USER}&password=${CHE_PASS}" ${TOKEN_ENDPOINT} | jq -r '.access_token')
}

# Teardown code after each test
teardown() {
    time_after=$SECONDS
    echo "# Test time taken: " $(($time_after-$time_before)) " seconds" >&3
}

@test "Codewind Sidecar Test 1: Create Che workspace from Codewind dev file" {
    deleteExistingCodewindCheWorkspaces
    createCodewindCheWorkspace
}
