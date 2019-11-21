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


# Create and start up a Che workspace based on the Codewind dev file
function createCodewindCheWorkspace() {
    if [ -f che_workspace_id.txt ]; then
        rm che_workspace_id.txt
    fi

    # Create Che workspace based on latest Codewind .yaml devfile converted to json
    local HTTP_RESPONSE=$(curl $CODEWIND_DEVFILE_URL | curl --silent --write-out "HTTPSTATUS:%{http_code}" --request POST --header 'Authorization: Bearer '"$CHE_ACCESS_TOKEN"'' --header "Content-Type:text/yaml" --data-binary @- $CHE_INGRESS_DOMAIN_URL/api/workspace/devfile?start-after-create=true)

    local HTTP_BODY=$(echo $HTTP_RESPONSE | sed -e 's/HTTPSTATUS\:.*//g')
    local HTTP_STATUS=$(echo $HTTP_RESPONSE | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

    if [[ $HTTP_STATUS != 201 ]]; then
        echo "# Error creating Che Codewind workspace [HTTP status: $HTTP_STATUS]" >&3
        exit 1
    fi

    # Write workspace ID to a file so that other bats tests can discover it
    echo $HTTP_BODY | jq -r .id > che_workspace_id.txt
}

# Stop the Codewind Che workspace
function stopCodewindCheWorkspace() {
    local HTTP_RESPONSE=$(curl --silent --header 'Authorization: Bearer '"$CHE_ACCESS_TOKEN"'' --write-out "HTTPSTATUS:%{http_code}" --request DELETE $CHE_INGRESS_DOMAIN_URL/api/workspace/$CHE_WORKSPACE_ID/runtime)

    local HTTP_BODY=$(echo $HTTP_RESPONSE | sed -e 's/HTTPSTATUS\:.*//g')
    local HTTP_STATUS=$(echo $HTTP_RESPONSE | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

    if [[ $HTTP_STATUS != 204 ]]; then
        echo "# Error stopping Che Codewind workspace [HTTP status: $HTTP_STATUS]" >&3
        exit 1
    fi

    # Wait 30 seconds for workspace to stop since the api operation is asynchronous
    sleep 30 
}

# Delete the Codewind Che workspace
function deleteCodewindCheWorkspace() {
   local HTTP_RESPONSE=$(curl --silent --header 'Authorization: Bearer '"$CHE_ACCESS_TOKEN"'' --write-out "HTTPSTATUS:%{http_code}" --request DELETE $CHE_INGRESS_DOMAIN_URL/api/workspace/$CHE_WORKSPACE_ID)

   local HTTP_BODY=$(echo $HTTP_RESPONSE | sed -e 's/HTTPSTATUS\:.*//g')
   local HTTP_STATUS=$(echo $HTTP_RESPONSE | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

   if [[ $HTTP_STATUS != 204 ]]; then
       echo "# Error deleting Che Codewind workspace [HTTP status: $HTTP_STATUS]" >&3
       exit 1
   fi
}

# Delete any existing Codewind Che workspaces
function deleteExistingCodewindCheWorkspaces() {
    # Get all Che Workspace IDs
    local HTTP_RESPONSE=$(curl --silent --header 'Authorization: Bearer '"$CHE_ACCESS_TOKEN"'' --write-out "HTTPSTATUS:%{http_code}" --request GET $CHE_INGRESS_DOMAIN_URL/api/workspace)
    local HTTP_BODY=$(echo $HTTP_RESPONSE | sed -e 's/HTTPSTATUS\:.*//g')
    local HTTP_STATUS=$(echo $HTTP_RESPONSE | tr -d '\n' | sed -e 's/.*HTTPSTATUS://')

    if [[ $HTTP_STATUS = 200 ]]; then
        # Get all Codewind Che workspaces IDs & iterate through them trying to stop and delete them
        CHE_WORKSPACE_IDS=$(echo $HTTP_BODY | jq -r '.[] | select(.devfile.metadata.name=="codewind-che") | .id')
        for i in ${CHE_WORKSPACE_IDS[@]}; do
            echo -e "# ${BLUE}Codewind Workspace ID: ${i} ${RESET}\n" >&3

            # Stop the Codewind Workspace
            echo -e "# ${BLUE}Stopping the Codewind Workspace ${RESET}\n" >&3
            HTTPSTATUS=$(curl -I --header 'Authorization: Bearer '"$CHE_ACCESS_TOKEN"'' --request DELETE $CHE_INGRESS_DOMAIN_URL/api/workspace/${i}/runtime 2>/dev/null | head -n 1 | cut -d$' ' -f2)
            if [[ $HTTPSTATUS -ne 204 ]]; then
                echo -e "# ${RED}Codewind workspace has failed to stop or is already stopped. Will attempt to remove the workspace... ${RESET}\n" >&3
            fi
            # Wait for the workspace to stop before removing it, otherwise the workspace removal fails
            echo -e "# ${BLUE}Sleeping for 10s to allow the workspace to stop before removing it ${RESET}\n"
            sleep 10

            # Remove the Codewind Workspace
            echo -e "# ${BLUE}Removing the Codewind Workspace ${RESET}\n" >&3
            HTTPSTATUS=$(curl -I --header 'Authorization: Bearer '"$CHE_ACCESS_TOKEN"'' --request DELETE $CHE_INGRESS_DOMAIN_URL/api/workspace/${i} 2>/dev/null | head -n 1 | cut -d$' ' -f2)
            if [[ $HTTPSTATUS -ne 204 ]]; then
                echo -e "# ${RED}Codewind workspace has failed to be removed... ${RESET}\n" >&3
                exit 1
            fi

            echo -e "# ${GREEN}Codewind should be removed momentarily... ${RESET}\n" >&3
        done
    fi
}

# Check for Codewind pod
function getCodewindPod {
    kubectl get pods --selector=app=codewind-pfe --no-headers $KUBE_NAMESPACE_ARG | grep $CHE_WORKSPACE_ID
}

# Get PID of process within a pod's container
# Arguments:
#       $1 (required): pod full name
#       $2 (required): container full name
#       $3 (required): process name
function getPIDofProcessInContainer {
    kubectl exec -t $1 $KUBE_NAMESPACE_ARG --container $2 -- pidof $3
}

# Examine sidecar container logs for specific filewatcher daemon messages to indicate successful start
# Arguments: 
#       $1 (optional): # of seconds to check in recent history of logs (if unset check entire log history)
function checkFilewatcherDaemonRunning {
    if [ ! -z "$1" ]; then
        since_arg=--since="$time_elapsed"s
    fi

    # Check sidecar logs for specific filewatcher daemon messages indicating successful start
    kubectl logs $CHE_WORKSPACE_POD_FULLNAME $SIDECAR_CONTAINER_FULLNAME $KUBE_NAMESPACE_ARG $since_arg | grep -E "Successfully connected to w(s){1,2}\:\/\/"
    kubectl logs $CHE_WORKSPACE_POD_FULLNAME $SIDECAR_CONTAINER_FULLNAME $KUBE_NAMESPACE_ARG $since_arg | grep -E "GET request completed, for http(s){0,1}\:\/\/"
}

# Check sidecar container for filewatcher daemon process
function getFileWatcherDaemonProcess {
    kubectl exec -t $CHE_WORKSPACE_POD_FULLNAME $KUBE_NAMESPACE_ARG --container $SIDECAR_CONTAINER_FULLNAME -- ps aux | grep filewatcherd
}

# Check if sidecar container is started and in ready state
# Arguments:
#       $1 (optional): # of restarts the container should have as a minimum (if unset don't check restart count)
function checkSidecarContainerReady {
     # Examine kubernetes pod metadata for sidecar state
    container_ready=$(kubectl get pods $CHE_WORKSPACE_POD_FULLNAME -o jsonpath="{.status.containerStatuses[?(@.name==\"$SIDECAR_CONTAINER_FULLNAME\")].ready}" $KUBE_NAMESPACE_ARG)
    container_restarts=$(kubectl get pods $CHE_WORKSPACE_POD_FULLNAME -o jsonpath="{.status.containerStatuses[?(@.name==\"$SIDECAR_CONTAINER_FULLNAME\")].restartCount}" $KUBE_NAMESPACE_ARG)

    [ $container_ready = "true" ]

    if [ ! -z "$1" ]; then
        (("$container_restarts" >= "$1"))
    fi
}
